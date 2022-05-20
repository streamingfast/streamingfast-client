package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/viper"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	sf "github.com/streamingfast/streamingfast-client"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var retryDelay = 5 * time.Second

type streamConfig struct {
	writer      io.Writer
	stats       *stats
	brange      BlockRange
	filter      string
	cursor      string
	endpoint    string
	handleForks bool
	transforms  []*anypb.Any
}

type protocolBlockFactory func() proto.Message
type protoToRef func(message proto.Message) sf.BlockRef

func newStream(endpoint string) (stream pbfirehose.StreamClient, client dfuse.Client, skipAuth bool, err error) {
	var clientOptions []dfuse.ClientOption
	apiKey := os.Getenv("STREAMINGFAST_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("SF_API_KEY")
		if apiKey == "" {
			clientOptions = []dfuse.ClientOption{dfuse.WithoutAuthentication()}
			skipAuth = true
		}
	}

	if viper.GetBool("global-skip-auth") {
		clientOptions = []dfuse.ClientOption{dfuse.WithoutAuthentication()}
		skipAuth = true
	}

	if authEndpoint := viper.GetString("global-auth-endpoint"); authEndpoint != "" && !skipAuth {
		clientOptions = []dfuse.ClientOption{dfuse.WithAuthURL(authEndpoint)}
	}

	client, err = dfuse.NewClient(endpoint, apiKey, clientOptions...)
	if err != nil {
		return nil, nil, false, fmt.Errorf("unable to create streamingfast client")
	}

	useInsecureTSLConnection := viper.GetBool("global-insecure")
	usePlainTextConnection := viper.GetBool("global-plaintext")

	if useInsecureTSLConnection && usePlainTextConnection {
		return nil, nil, false, fmt.Errorf("option --insecure and --plaintext are mutually exclusive, they cannot be both specified at the same time")
	}

	var dialOptions []grpc.DialOption
	switch {
	case usePlainTextConnection:
		zlog.Debug("Configuring transport to use a plain text connection")
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	case useInsecureTSLConnection:
		zlog.Debug("Configuring transport to use an insecure TLS connection (skips certificate verification)")
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))}
	}

	conn, err := dgrpc.NewExternalClient(endpoint, dialOptions...)
	if err != nil {
		return nil, nil, false, fmt.Errorf("unable to create external gRPC client")
	}

	return pbfirehose.NewStreamClient(conn), client, skipAuth, err
}

func launchStream(ctx context.Context, config streamConfig, blkFactory protocolBlockFactory, toRef protoToRef) error {
	nextStatus := time.Now().Add(statusFrequency)
	cursor := config.cursor
	lastBlockRef := sf.EmptyBlockRef

	zlog.Info("starting stream",
		zap.Stringer("range", config.brange),
		zap.String("cursor", config.cursor),
		zap.String("endpoint", config.endpoint),
		zap.Bool("handle_forks", config.handleForks),
	)

	firehoseClient, dfuse, skipAuth, err := newStream(config.endpoint)
	if err != nil {
		return err
	}

stream:
	for {
		grpcCallOpts := []grpc.CallOption{}

		if !skipAuth {
			tokenInfo, err := dfuse.GetAPITokenInfo(ctx)
			if err != nil {
				return fmt.Errorf("unable to retrieve StreamingFast API token: %w", err)
			}
			credentials := oauth.NewOauthAccess(&oauth2.Token{AccessToken: tokenInfo.Token, TokenType: "Bearer"})

			grpcCallOpts = append(grpcCallOpts, grpc.PerRPCCredentials(credentials))
		}

		forkSteps := []pbfirehose.ForkStep{pbfirehose.ForkStep_STEP_NEW}
		if config.handleForks {
			forkSteps = append(forkSteps, pbfirehose.ForkStep_STEP_IRREVERSIBLE, pbfirehose.ForkStep_STEP_UNDO)
		}

		request := &pbfirehose.Request{
			StartBlockNum:     config.brange.Start,
			StartCursor:       config.cursor,
			StopBlockNum:      config.brange.End,
			ForkSteps:         forkSteps,
			IncludeFilterExpr: config.filter,
			Transforms:        config.transforms,
		}

		zlog.Debug("Initiating stream with remote endpoint", zap.String("endpoint", config.endpoint))
		stream, err := firehoseClient.Blocks(context.Background(), request, grpcCallOpts...)
		if err != nil {
			return fmt.Errorf("unable to start blocks stream: %w", err)
		}

		for {
			zlog.Debug("Waiting for message to reach us")
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break stream
				}

				zlog.Error("Stream encountered a remote error, going to retry",
					zap.String("cursor", cursor),
					zap.Stringer("last_block", lastBlockRef),
					zap.Duration("retry_delay", retryDelay),
					zap.Error(err),
				)
				break
			}

			zlog.Debug("Decoding received message's block")
			block := blkFactory()
			if err = anypb.UnmarshalTo(response.Block, block, proto.UnmarshalOptions{}); err != nil {
				return fmt.Errorf("should have been able to unmarshal received block payload: %w", err)
			}

			cursor = response.Cursor
			lastBlockRef = toRef(block)

			if tracer.Enabled() {
				zlog.Debug("Block received",
					zap.Stringer("block", lastBlockRef),
					zap.String("cursor", cursor),
				)
			}

			now := time.Now()
			if now.After(nextStatus) {
				zlog.Info("Stream blocks progress",
					zap.Object("stats", config.stats),
				)
				nextStatus = now.Add(statusFrequency)
			}

			if config.writer != nil {
				if err := writeBlock(config.writer, response, lastBlockRef); err != nil {
					return err
				}
			}

			config.stats.recordBlock(int64(proto.Size(block)))
		}

		time.Sleep(5 * time.Second)
		config.stats.restartCount.IncBy(1)
	}

	elapsed := config.stats.duration()

	println("")
	println("Completed streaming")
	printf("Duration: %s\n", elapsed)
	printf("Time to first block: %s\n", config.stats.timeToFirstBlock)
	if config.stats.restartCount.total > 0 {
		printf("Restart count: %s\n", config.stats.restartCount.Overall(elapsed))
	}

	println("")
	printf("Block received: %s\n", config.stats.blockReceived.Overall(elapsed))
	printf("Bytes received: %s\n", config.stats.bytesReceived.Overall(elapsed))
	return nil
}
