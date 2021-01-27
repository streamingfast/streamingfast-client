package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dfuse-io/bstream"
	dfuse "github.com/dfuse-io/client-go"
	"github.com/dfuse-io/dgrpc"
	"github.com/dfuse-io/jsonpb"
	"github.com/dfuse-io/logging"
	pbbstream "github.com/dfuse-io/pbgo/dfuse/bstream/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/paulbellamy/ratecounter"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/dfuse/ethereum/codec/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

var retryDelay = 5 * time.Second
var statusFrequency = 15 * time.Second
var traceEnabled = logging.IsTraceEnabled("consumer", "github.com/streamingfast/streamingfast-client")
var zlog = logging.NewSimpleLogger("consumer", "github.com/streamingfast/streamingfast-client")

var flagInsecure = flag.Bool("i", false, "When set, assume with talk over a plain-text unecrypted gRPC connection")
var flagSkipVerify = flag.Bool("s", false, "When set, skips certification verification")
var flagWrite = flag.String("o", "-", "When set, write each block as one JSON line in the specified file, value '-' writes to standard output otherwise to a file, {range} is replaced by block range in this case")
var flagStartCursor = flag.String("start-cursor", "", "Last cursor used to continue where you left off")

func main() {
	flag.Parse()

	// [filter query] [start_block] [end_block]
	// [filter query] [start_block]
	// [filter-query] -start-cursor [cursor]

	args := flag.Args()
	ensure((len(args) == 1 && *flagStartCursor != "") || len(args) > 1, errorUsage("missing arguments"))

	filter := args[0]
	cursor := *flagStartCursor
	var brange blockRange
	if cursor == "" {
		brange = newBlockRange(args[1:])
	}

	apiKey := os.Getenv("STREAMINGFAST_API_KEY")
	ensure(apiKey != "", errorUsage("the environment variable STREAMINGFAST_API_KEY must be set to a valid dfuse API key value"))

	var dialOptions []grpc.DialOption
	if *flagSkipVerify {
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))}
	}

	dfuse, err := dfuse.NewClient("mainnet.eth.dfuse.io", apiKey)
	noError(err, "unable to create dfuse client")

	conn, err := dgrpc.NewExternalClient("blocks.mainnet.eth.dfuse.io:443", dialOptions...)
	noError(err, "unable to create external gRPC client")

	streamClient := pbbstream.NewBlockStreamV2Client(conn)

	stats := newStats()
	nextStatus := time.Now().Add(statusFrequency)
	writer, closer := blockWriter(brange)
	defer closer()

	lastBlockRef := bstream.BlockRefEmpty

	zlog.Info("Starting stream", zap.Stringer("range", brange), zap.String("cursor", cursor))
stream:
	for {
		tokenInfo, err := dfuse.GetAPITokenInfo(context.Background())
		noError(err, "unable to retrieve StreamingFast API token")

		credentials := oauth.NewOauthAccess(&oauth2.Token{AccessToken: tokenInfo.Token, TokenType: "Bearer"})
		stream, err := streamClient.Blocks(context.Background(), &pbbstream.BlocksRequestV2{
			StartBlockNum:     brange.start,
			StartCursor:       cursor,
			StopBlockNum:      brange.end,
			ForkSteps:         []pbbstream.ForkStep{pbbstream.ForkStep_STEP_IRREVERSIBLE, pbbstream.ForkStep_STEP_NEW, pbbstream.ForkStep_STEP_UNDO},
			IncludeFilterExpr: filter,
			Details:           pbbstream.BlockDetails_BLOCK_DETAILS_FULL,
		}, grpc.PerRPCCredentials(credentials))
		noError(err, "unable to start blocks stream")

		for {
			zlog.Debug("Waiting for message to reach us")
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break stream
				}

				zlog.Error("Stream encountered a remote error, going to retry", zap.String("cursor", cursor), zap.Stringer("last_block", lastBlockRef), zap.Duration("retry_delay", retryDelay), zap.Error(err))
				break
			}

			zlog.Debug("Decoding received message's block")
			block := &pbcodec.Block{}
			err = ptypes.UnmarshalAny(response.Block, block)
			noError(err, "should have been able to unmarshal received block payload")

			cursor = response.Cursor
			lastBlockRef = block.AsRef()

			if traceEnabled {
				zlog.Debug("Block received", zap.Stringer("block", lastBlockRef), zap.Stringer("previous", bstream.NewBlockRefFromID(block.PreviousID())), zap.String("cursor", cursor))
			}

			now := time.Now()
			if now.After(nextStatus) {
				zlog.Info("Stream blocks progress", zap.Object("stats", stats))
				nextStatus = now.Add(statusFrequency)
			}

			if writer != nil {
				writeBlock(writer, response, block)
			}

			stats.recordBlock(int64(response.XXX_Size()))
		}

		time.Sleep(5 * time.Second)
		stats.restartCount.IncBy(1)
	}

	elapsed := stats.duration()

	println("")
	println("Completed streaming")
	printf("Duration: %s\n", elapsed)
	printf("Time to first block: %s\n", stats.timeToFirstBlock)
	if stats.restartCount.total > 0 {
		printf("Restart count: %s\n", stats.restartCount.Overall(elapsed))
	}

	println("")
	printf("Block received: %s\n", stats.blockReceived.Overall(elapsed))
	printf("Bytes received: %s\n", stats.bytesReceived.Overall(elapsed))
}

var endOfLine = []byte("\n")

func writeBlock(writer io.Writer, response *pbbstream.BlockResponseV2, block *pbcodec.Block) {
	line, err := jsonpb.MarshalToString(response)
	noError(err, "unable to marshal block %s to JSON", block.AsRef())

	_, err = writer.Write([]byte(line))
	noError(err, "unable to write block %s line to JSON", block.AsRef())

	_, err = writer.Write(endOfLine)
	noError(err, "unable to write block %s line ending", block.AsRef())
}

func blockWriter(bRange blockRange) (io.Writer, func()) {
	if flagWrite == nil || strings.TrimSpace(*flagWrite) == "" {
		return nil, func() {}
	}

	out := strings.Replace(strings.TrimSpace(*flagWrite), "{range}", strings.ReplaceAll(bRange.String(), " ", ""), 1)
	if out == "-" {
		return os.Stdout, func() {}
	}

	dir := filepath.Dir(out)
	noError(os.MkdirAll(dir, os.ModePerm), "unable to create directories %q", dir)

	file, err := os.Create(out)
	noError(err, "unable to create file %q", out)

	return file, func() { file.Close() }
}

type stats struct {
	startTime        time.Time
	timeToFirstBlock time.Duration
	blockReceived    *counter
	bytesReceived    *counter
	restartCount     *counter
}

func newStats() *stats {
	return &stats{
		startTime:     time.Now(),
		blockReceived: &counter{0, ratecounter.NewRateCounter(1 * time.Second), "block", "s"},
		bytesReceived: &counter{0, ratecounter.NewRateCounter(1 * time.Second), "byte", "s"},
		restartCount:  &counter{0, ratecounter.NewRateCounter(1 * time.Minute), "restart", "m"},
	}
}

func (s *stats) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("block", s.blockReceived.String())
	encoder.AddString("bytes", s.bytesReceived.String())
	return nil
}

func (s *stats) duration() time.Duration {
	return time.Now().Sub(s.startTime)
}

func (s *stats) recordBlock(payloadSize int64) {
	if s.timeToFirstBlock == 0 {
		s.timeToFirstBlock = time.Now().Sub(s.startTime)
	}

	s.blockReceived.IncBy(1)
	s.bytesReceived.IncBy(payloadSize)
}

// arg"11700000 - 11700001"
// -1000

func newBlockRange(args []string) (out blockRange) {
	ensure(isInt(args[0]), "the <range> start value %q is not a valid uint64 value", args[0])
	out.start, _ = strconv.ParseInt(args[0], 10, 64)
	if len(args) == 1 {
		return
	}
	ensure(isUint(args[1]), "the <range> end value %q is not a valid uint64 value", args[1])
	out.end, _ = strconv.ParseUint(args[1], 10, 64)
	ensure(out.start < int64(out.end), "the <range> start value %q value comes after end value %q", args[0], args[1])
	return
}

func isUint(in string) bool {
	_, err := strconv.ParseUint(in, 10, 64)
	return err == nil
}

func isInt(in string) bool {
	_, err := strconv.ParseInt(in, 10, 64)
	return err == nil
}

func errorUsage(message string, args ...interface{}) string {
	return fmt.Sprintf(message+"\n\n"+usage(), args...)
}

func usage() string {
	return `usage: sf <filter> [start_block] [end_block]

Other invocations:

    sf <filter> [start_block] [end_block]    # Stream a given block range
    sf <filter> [start_block]                # Stream real-time blocks
	sf --start-cursor [cursor] <filter>       # Stream from where you left off


<filter> is a valid CEL filter expression for the Ethereum network.

start_block - can be positive (an absolute reference to a block), or
              negative (a number of blocks from the tip of the chain).

end_block - an absolute end block, if not specified, the stream will stop
            when the Ethereum network stops: never.
`
}

func ensure(condition bool, message string, args ...interface{}) {
	if !condition {
		noError(fmt.Errorf(message, args...), "invalid arguments")
	}
}

func noError(err error, message string, args ...interface{}) {
	if err != nil {
		quit(message+": "+err.Error(), args...)
	}
}

func quit(message string, args ...interface{}) {
	printf(message+"\n", args...)
	os.Exit(1)
}

func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func println(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

type blockRange struct {
	start int64
	end   uint64
}

func (b blockRange) String() string {
	return fmt.Sprintf("%d - %d", b.start, b.end)
}

type counter struct {
	total    uint64
	counter  *ratecounter.RateCounter
	unit     string
	timeUnit string
}

func (c *counter) IncBy(value int64) {
	if value <= 0 {
		return
	}

	c.counter.Incr(value)
	c.total += uint64(value)
}

func (c *counter) Total() uint64 {
	return c.total
}

func (c *counter) Rate() int64 {
	return c.counter.Rate()
}

func (c *counter) String() string {
	return fmt.Sprintf("%d %s/%s (%d total)", c.counter.Rate(), c.unit, c.timeUnit, c.total)
}

func (c *counter) Overall(elapsed time.Duration) string {
	rate := float64(c.total)
	if elapsed.Minutes() > 1 {
		rate = rate / elapsed.Minutes()
	}

	return fmt.Sprintf("%d %s/%s (%d %s total)", uint64(rate), c.unit, "min", c.total, c.unit)
}
