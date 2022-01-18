package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/sf/solana/codec/v1"
	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

var solSfCmd = &cobra.Command{
	Use: "sol [flags] [<start_block>] [<end_block>]",
	Short: `Streaming Fast Ethereum client
usage: sf sol [flags] [<start_block>] [<end_block>]
	`,
	Long: usage,
	Args: cobra.MinimumNArgs(0),
	RunE: solSfCmdE,
}

func init() {
	RootCmd.AddCommand(solSfCmd)

	// Transforms
	solSfCmd.Flags().String("program-filter", "", "Specifcy a comma delimited base58 program addresses to filter a block for given programs")
}

func solSfCmdE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	startCursor := viper.GetString("global-start-cursor")
	endpoint := viper.GetString("global-endpoint")
	outputFlag := viper.GetString("global-output")

	inputs, err := checkArgs(startCursor, args)
	if err != nil {
		return err
	}

	if e := os.Getenv("STREAMINGFAST_ENDPOINT"); e != "" {
		endpoint = e
	}

	if endpoint == "" {
		return fmt.Errorf("unable to resolve endpoint")
	}

	var clientOptions []dfuse.ClientOption
	apiKey := os.Getenv("STREAMINGFAST_API_KEY")
	if apiKey == "" {
		clientOptions = []dfuse.ClientOption{dfuse.WithoutAuthentication()}
	}

	dfuse, err := dfuse.NewClient(endpoint, apiKey, clientOptions...)
	if err != nil {
		return fmt.Errorf("unable to create streamingfast client")
	}

	var dialOptions []grpc.DialOption
	if viper.GetBool("global-insecure") {
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))}
	}

	conn, err := dgrpc.NewExternalClient(endpoint, dialOptions...)
	if err != nil {
		return fmt.Errorf("unable to create external gRPC client")
	}

	writer, closer, err := blockWriter(inputs.Brange, outputFlag)
	if err != nil {
		return fmt.Errorf("unable to setup writer: %w", err)
	}
	defer closer()

	return launchStream(ctx, streamConfig{
		client:      pbfirehose.NewStreamClient(conn),
		dfuseCli:    dfuse,
		writer:      writer,
		stats:       newStats(),
		brange:      inputs.Brange,
		cursor:      startCursor,
		endpoint:    endpoint,
		handleForks: viper.GetBool("global-handle-forks"),
		skipAuth:    viper.GetBool("global-skip-auth"),
	},
		func() proto.Message {
			return &pbcodec.Block{}
		},
		func(message proto.Message) bstream.BlockRef {
			return message.(*pbcodec.Block).AsRef()
		})
}
