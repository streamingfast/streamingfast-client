package cmd

import (
	"fmt"
	"github.com/paulbellamy/ratecounter"
	"github.com/streamingfast/jsonpb"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/sf/ethereum/codec/v1"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var statusFrequency = 15 * time.Second

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

type blockRange struct {
	start int64
	end   uint64
}

func newBlockRange(startBlockNum, stopBlockNum string) (blockRange, error) {
	if !isInt(startBlockNum) {
		return blockRange{}, fmt.Errorf("the <range> start value %q is not a valid uint64 value", startBlockNum)
	}
	out := blockRange{}
	out.start, _ = strconv.ParseInt(startBlockNum, 10, 64)
	if stopBlockNum == "" {
		return out, nil
	}
	if !isUint(stopBlockNum) {
		return blockRange{}, fmt.Errorf("the <range> end value %q is not a valid uint64 value", stopBlockNum)
	}
	out.end, _ = strconv.ParseUint(stopBlockNum, 10, 64)
	if out.start > int64(out.end) {
		return blockRange{}, fmt.Errorf("the <range> start value %q value comes after end value %q", startBlockNum, stopBlockNum)
	}
	return out, nil
}

func (b blockRange) String() string {
	return fmt.Sprintf("%d - %d", b.start, b.end)
}

func blockWriter(bRange blockRange, flagWrite string) (io.Writer, func(), error) {
	if strings.TrimSpace(flagWrite) == "" {
		return nil, func() {}, nil
	}

	out := strings.Replace(strings.TrimSpace(flagWrite), "{range}", strings.ReplaceAll(bRange.String(), " ", ""), 1)
	if out == "-" {
		return os.Stdout, func() {}, nil
	}

	dir := filepath.Dir(out)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to create directories %q", dir)
	}

	file, err := os.Create(out)
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to create file %q", out)
	}
	return file, func() { file.Close() }, nil
}

var endOfLine = []byte("\n")

func writeBlock(writer io.Writer, response *pbfirehose.Response, block *pbcodec.Block) error {
	line, err := jsonpb.MarshalToString(response)
	if err != nil {
		return fmt.Errorf("unable to marshal block %s to JSON", block.AsRef())
	}

	_, err = writer.Write([]byte(line))
	if err != nil {
		return fmt.Errorf("unable to write block %s line to JSON", block.AsRef())
	}

	_, err = writer.Write(endOfLine)
	if err != nil {
		return fmt.Errorf("unable to write block %s line ending", block.AsRef())
	}
	return nil
}
