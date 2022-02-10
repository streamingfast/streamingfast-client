package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/streamingfast/eth-go"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/sf/ethereum/codec/v1"
	pbtransforms "github.com/streamingfast/streamingfast-client/pb/sf/ethereum/transforms/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var ethSfCmd = &cobra.Command{
	Use:   "eth [flags] [<start_block>] [<end_block>]",
	Short: `StreamingFast Ethereum client`,
	Long:  usage,
	Args:  cobra.MaximumNArgs(2),
	RunE:  ethSfRunE,
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
	ethSfCmd.Flags().Bool("light-block", false, "When set, returned blocks will be stripped of some information")
	ethSfCmd.Flags().Bool("log-filter", false, "When set, returned blocks will have log filters applied")
	ethSfCmd.Flags().StringSlice("log-filter-addresses", nil, "List of addresses to filter blocks with")
	ethSfCmd.Flags().StringSlice("log-filter-event-sigs", nil, "List of event signatures to filter blocks with")
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
	skipAuth := viper.GetBool("global-skip-auth")

	inputs, err := checkArgs(startCursor, args)
	if err != nil {
		return err
	}
	zlog.Debug("arguments", zap.Reflect("args", inputs))

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

	transforms := []*anypb.Any{}

	if viper.GetBool("eth-cmd-light-block") {
		t, err := lightBlockTransform()
		if err != nil {
			return fmt.Errorf("unable to create light block transform: %w", err)
		}
		transforms = append(transforms, t)
	}

	if viper.GetBool("eth-cmd-log-filter") {
		var addrs []eth.Address
		var sigs []eth.Hash
		addrStrings := viper.GetStringSlice("eth-cmd-log-filter-addresses")
		sigStrings := viper.GetStringSlice("eth-cmd-log-filter-event-sigs")

		for _, addrString := range addrStrings {
			addr := eth.MustNewAddress(addrString)
			addrs = append(addrs, addr)
		}

		for _, sigString := range sigStrings {
			sig := eth.MustNewHash(sigString)
			sigs = append(sigs, sig)
		}

		t, err := logFilterTransform(addrs, sigs)
		if err != nil {
			return fmt.Errorf("unable to create log filter transform: %w", err)
		}

		transforms = append(transforms, t)
	}

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
		transforms:  transforms,
	},
		func() proto.Message {
			return &pbcodec.Block{}
		},
		func(message proto.Message) bstream.BlockRef {
			return message.(*pbcodec.Block).AsRef()
		})
}

func lightBlockTransform() (*anypb.Any, error) {
	transform := &pbtransforms.LightBlock{}
	return anypb.New(transform)
}

func logFilterTransform(addrs []eth.Address, sigs []eth.Hash) (*anypb.Any, error) {
	var addrBytes [][]byte
	var sigsBytes [][]byte

	for _, addr := range addrs {
		b := addr.Bytes()
		addrBytes = append(addrBytes, b)
	}

	for _, sig := range sigs {
		b := sig.Bytes()
		sigsBytes = append(sigsBytes, b)
	}

	transform := &pbtransforms.BasicLogFilter{
		Addresses:       addrBytes,
		EventSignatures: sigsBytes,
	}

	return anypb.New(transform)
}
