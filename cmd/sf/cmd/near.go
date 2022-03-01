package cmd

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	sf "github.com/streamingfast/streamingfast-client"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/sf/near/codec/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

var nearSfCmd = &cobra.Command{
	Use:   "near [flags] [<start_block>] [<end_block>]",
	Short: `StreamingFast Near client`,
	Long:  usage,
	RunE:  nearSfCmdE,
}

func init() {
	RootCmd.AddCommand(nearSfCmd)

	defaultNearEndpoint := "mainnet.near.streamingfast.io:443"
	if e := os.Getenv("STREAMINGFAST_ENDPOINT"); e != "" {
		defaultNearEndpoint = e
	}
	nearSfCmd.Flags().StringP("endpoint", "e", defaultNearEndpoint, "The endpoint to connect the stream of blocks (default value set by STREAMINGFAST_ENDPOINT env var, can be overriden by network-specific flags like --testnet)")
	nearSfCmd.Flags().Bool("testnet", false, "When set, will switch default endpoint testnet")
}

func nearSfCmdE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	startCursor := viper.GetString("global-start-cursor")
	endpoint := viper.GetString("near-cmd-endpoint")
	outputFlag := viper.GetString("global-output")
	skipAuth := viper.GetBool("global-skip-auth")

	testnet := viper.GetBool("near-cmd-testnet")

	inputs, err := checkArgs(startCursor, args)
	if err != nil {
		return err
	}

	if testnet {
		endpoint = "testnet.near.streamingfast.io:443"
	}
	if endpoint == "" {
		return fmt.Errorf("unable to resolve endpoint")
	}

	var clientOptions []dfuse.ClientOption
	apiKey := os.Getenv("STREAMINGFAST_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("SF_API_KEY")
		if apiKey == "" {
			clientOptions = []dfuse.ClientOption{dfuse.WithoutAuthentication()}
			skipAuth = true
		}
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

	writer, closer, err := blockWriter(inputs.Range, outputFlag)
	if err != nil {
		return fmt.Errorf("unable to setup writer: %w", err)
	}
	defer closer()

	return launchStream(ctx, streamConfig{
		client:      pbfirehose.NewStreamClient(conn),
		dfuseCli:    dfuse,
		writer:      writer,
		stats:       newStats(),
		brange:      inputs.Range,
		cursor:      startCursor,
		endpoint:    endpoint,
		handleForks: viper.GetBool("global-handle-forks"),
		skipAuth:    skipAuth,
	},
		func() proto.Message {
			return &pbcodec.Block{}
		},
		func(message proto.Message) sf.BlockRef {
			b := message.(*pbcodec.Block)
			return sf.NewBlockRef(b.Header.Hash.String(), b.Header.Height)
		})
}
