package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	pbtransforms "github.com/streamingfast/streamingfast-client/pb/sf/ethereum/transforms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/anypb"
	"os"
)

var ethSfCmd = &cobra.Command{
	Use: "eth [flags] <filter> [<start_block>] [<end_block>]",
	Short: `Streaming Fast Ethereum client
usage: sf eth [flags] <filter> [<start_block>] [<end_block>]
	`,
	Long: usage,
	Args: cobra.MinimumNArgs(1),
	RunE: ethSfRunE,
}

func init() {
	RootCmd.AddCommand(ethSfCmd)

	// Endpoint settings
	ethSfCmd.Flags().Bool("bsc", false, "When set, will force the endpoint to Binance Smart Chain")
	ethSfCmd.Flags().Bool("polygon", false, "When set, will force the endpoint to Polygon (previously Matic)")
	ethSfCmd.Flags().Bool("heco", false, "When set, will force the endpoint to Huobi Eco Chain")
	ethSfCmd.Flags().Bool("fantom", false, "When set, will force the endpoint to Fantom Opera Mainnet")
	ethSfCmd.Flags().Bool("xdai", false, "When set, will force the endpoint to xDai Chain")

	// Transforms
	ethSfCmd.Flags().Bool("light-block", false, "When set returned blocks will be stripped of some information")
}

func ethSfRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	startCursor := viper.GetString("global-start-cursor")
	endpoint := viper.GetString("global-endpoint")
	bscNetwork := viper.GetBool("eth-cmd-bsc")
	polygonNetwork := viper.GetBool("eth-cmd-polygon")
	hecoNetwork := viper.GetBool("eth-cmd-heco")
	fantomNetwork := viper.GetBool("eth-cmd-fantom")
	xdaiNetwork := viper.GetBool("eth-cmd-xdai")
	outputFlag := viper.GetString("global-output")

	inputs, err := checkArgs(startCursor, args)
	if err != nil {
		return err
	}

	if !noMoreThanOneTrue(bscNetwork, polygonNetwork, hecoNetwork, fantomNetwork, xdaiNetwork) {
		return fmt.Errorf("cannot set more than one network flag (ex: --polygon, --bsc)")
	}

	switch {
	case bscNetwork:
		endpoint = "bsc.streamingfast.io:443"
	case polygonNetwork:
		endpoint = "polygon.streamingfast.io:443"
	case hecoNetwork:
		endpoint = "heco.streamingfast.io:443"
	case fantomNetwork:
		endpoint = "fantom.streamingfast.io:443"
	case xdaiNetwork:
		endpoint = "xdai.streamingfast.io:443"
	default:
		if e := os.Getenv("STREAMINGFAST_ENDPOINT"); e != "" {
			endpoint = e
		}
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

	writer, closer, err := blockWriter(inputs.brange, outputFlag)
	if err != nil {
		return fmt.Errorf("unable to setup writer: %w", err)
	}
	defer closer()

	transforms := []*anypb.Any{}
	if viper.GetBool("eth-cmd-light-block") {
		tranform, err := lightBlockTransform()
		if err != nil {
			return fmt.Errorf("unable to create light nlock transfor: %w", err)
		}
		transforms = append(transforms, tranform)
	}

	return launchStream(ctx, streamConfig{
		client:      pbfirehose.NewStreamClient(conn),
		dfuseCli:    dfuse,
		writer:      writer,
		stats:       newStats(),
		brange:      inputs.brange,
		filter:      inputs.filter,
		cursor:      startCursor,
		endpoint:    endpoint,
		handleForks: viper.GetBool("global-handle-forks"),
		skipAuth:    viper.GetBool("global-skip-auth"),
		transforms:  transforms,
	})
}

func lightBlockTransform() (*anypb.Any, error) {
	transform := &pbtransforms.LightBlock{}
	return anypb.New(transform)
}
