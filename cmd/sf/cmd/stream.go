package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/streamingfast/bstream"
	dfuse "github.com/streamingfast/client-go"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var retryDelay = 5 * time.Second

type streamConfig struct {
	client   pbfirehose.StreamClient
	dfuseCli dfuse.Client
	writer   io.Writer
	stats    *stats

	brange      BlockRange
	filter      string
	cursor      string
	endpoint    string
	handleForks bool
	skipAuth    bool
	transforms  []*anypb.Any
}

type protocolBlockFactory func() proto.Message
type protoToRef func(message proto.Message) bstream.BlockRef

func launchStream(ctx context.Context, config streamConfig, blkFactory protocolBlockFactory, toRef protoToRef) error {
	nextStatus := time.Now().Add(statusFrequency)
	cursor := config.cursor
	lastBlockRef := bstream.BlockRefEmpty
	zlog.Info("starting stream",
		zap.Stringer("range", config.brange),
		zap.String("cursor", config.cursor),
		zap.String("endpoint", config.endpoint),
		zap.Bool("handle_forks", config.handleForks),
	)
stream:
	for {
		grpcCallOpts := []grpc.CallOption{}

		if !config.skipAuth {
			tokenInfo, err := config.dfuseCli.GetAPITokenInfo(ctx)
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

		stream, err := config.client.Blocks(context.Background(), request, grpcCallOpts...)
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

			if traceEnabled {
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
