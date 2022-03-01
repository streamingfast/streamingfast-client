package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/streamingfast/eth-go"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	dfuse "github.com/streamingfast/client-go"
	"github.com/streamingfast/dgrpc"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v1"
	sf "github.com/streamingfast/streamingfast-client"
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

	defaultEthEndpoint := "api.streamingfast.io:443"
	if e := os.Getenv("STREAMINGFAST_ENDPOINT"); e != "" {
		defaultEthEndpoint = e
	}
	ethSfCmd.Flags().StringP("endpoint", "e", defaultEthEndpoint, "The endpoint to connect the stream of blocks (default value set by STREAMINGFAST_ENDPOINT env var, can be overriden by network-specific flags like --bsc)")

	// Endpoint settings
	ethSfCmd.Flags().Bool("bsc", false, "When set, will change the default endpoint to Binance Smart Chain")
	ethSfCmd.Flags().Bool("polygon", false, "When set, will change the default endpoint to Polygon (previously Matic)")
	ethSfCmd.Flags().Bool("heco", false, "When set, will change the default endpoint to Huobi Eco Chain")
	ethSfCmd.Flags().Bool("fantom", false, "When set, will change the default endpoint to Fantom Opera Mainnet")
	ethSfCmd.Flags().Bool("xdai", false, "When set, will change the default endpoint to xDai Chain")

	// Transforms
	ethSfCmd.Flags().Bool("light-block", false, "When set, returned blocks will be stripped of some information")
	ethSfCmd.Flags().StringSlice("log-filter-multi", nil, "Advanced filter. List of 'address[+address[+...]]:eventsig[+eventsig[+...]]' pairs, ex: 'dead+beef:1234+5678,:0x44,0x12:' results in 3 filters. Mutually exclusive with --log-filter-addresses and --log-filter-event-sigs.")
	ethSfCmd.Flags().StringSlice("log-filter-addresses", nil, "Basic filter. List of addresses with which to filter blocks. Mutually exclusive with --log-filter-multi.")
	ethSfCmd.Flags().StringSlice("log-filter-event-sigs", nil, "Basic filter. List of event signatures with which to filter blocks. Mutually exclusive with --log-filter-multi.")
}

func ethSfRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	startCursor := viper.GetString("global-start-cursor")

	endpoint := viper.GetString("eth-cmd-endpoint")
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

	multiFilter := viper.GetStringSlice("eth-cmd-log-filter-multi")
	addrFilters := viper.GetStringSlice("eth-cmd-log-filter-addresses")
	sigFilters := viper.GetStringSlice("eth-cmd-log-filter-event-sigs")

	hasMultiFilter := len(multiFilter) != 0
	hasSingleFilter := len(addrFilters) != 0 || len(sigFilters) != 0

	switch {
	case hasMultiFilter && hasSingleFilter:
		return fmt.Errorf("options --log-filter-{addresses|event-sigs} are incompatible with --log-filter-multi, don't use both")
	case hasMultiFilter:
		mft, err := parseMultiLogFilter(multiFilter)
		if err != nil {
			return err
		}
		if mft != nil {
			transforms = append(transforms, mft)
		}
		break
	case hasSingleFilter:

		t, err := parseSingleLogFilter(addrFilters, sigFilters)
		if err != nil {
			return fmt.Errorf("unable to create log filter transform: %w", err)
		}

		if t != nil {
			transforms = append(transforms, t)
		}
		break
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
		func(message proto.Message) sf.BlockRef {
			return message.(*pbcodec.Block).AsRef()
		})
}

func lightBlockTransform() (*anypb.Any, error) {
	transform := &pbtransforms.LightBlock{}
	return anypb.New(transform)
}

func basicLogFilter(addrs []eth.Address, sigs []eth.Hash) *pbtransforms.BasicLogFilter {
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

	return &pbtransforms.BasicLogFilter{
		Addresses:       addrBytes,
		EventSignatures: sigsBytes,
	}
}

func parseSingleLogFilter(addrFilters []string, sigFilters []string) (*anypb.Any, error) {
	var addrs []eth.Address
	for _, addrString := range addrFilters {
		addr := eth.MustNewAddress(addrString)
		addrs = append(addrs, addr)
	}

	var sigs []eth.Hash
	for _, sigString := range sigFilters {
		sig := eth.MustNewHash(sigString)
		sigs = append(sigs, sig)
	}

	transform := basicLogFilter(addrs, sigs)
	return anypb.New(transform)
}

func parseMultiLogFilter(in []string) (*anypb.Any, error) {

	mf := &pbtransforms.MultiLogFilter{}

	for _, filter := range in {
		parts := strings.Split(filter, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("option --log-filter-multi must be of type address_hash+address_hash+address_hash:event_sig_hash+event_sig_hash (repeated, separated by comma)")
		}
		var addrs []eth.Address
		for _, a := range strings.Split(parts[0], "+") {
			if a != "" {
				addr := eth.MustNewAddress(a)
				addrs = append(addrs, addr)
			}
		}
		var sigs []eth.Hash
		for _, s := range strings.Split(parts[1], "+") {
			if s != "" {
				sig := eth.MustNewHash(s)
				sigs = append(sigs, sig)
			}
		}

		mf.BasicLogFilters = append(mf.BasicLogFilters, basicLogFilter(addrs, sigs))
	}

	t, err := anypb.New(mf)
	if err != nil {
		return nil, err
	}
	return t, nil

}
