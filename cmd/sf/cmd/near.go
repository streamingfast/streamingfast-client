package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sf "github.com/streamingfast/streamingfast-client"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/sf/near/codec/v1"
	pbtransforms "github.com/streamingfast/streamingfast-client/pb/sf/near/transforms/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var nearSfCmd = &cobra.Command{
	Use:   "near [<start_block>] [<end_block>]",
	Short: `StreamingFast NEAR Client`,
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
	nearSfCmd.Flags().StringSlice("filter-accounts", nil, "Basic filter. List of accounts with which to filter blocks")
}

func nearSfCmdE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	startCursor := viper.GetString("global-start-cursor")
	endpoint := viper.GetString("near-cmd-endpoint")
	outputFlag := viper.GetString("global-output")
	testnet := viper.GetBool("near-cmd-testnet")
	filterAccounts := viper.GetStringSlice("near-cmd-filter-accounts")

	transforms := []*anypb.Any{}
	if len(filterAccounts) != 0 {
		t := &pbtransforms.BasicReceiptFilter{
			Accounts: filterAccounts,
		}
		transformAny, err := anypb.New(t)
		if err != nil {
			return fmt.Errorf("processing transform for basicReceiptFilter: %w", err)
		}
		transforms = append(transforms, transformAny)
	}

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

	writer, closer, err := blockWriter(inputs.Range, outputFlag)
	if err != nil {
		return fmt.Errorf("unable to setup writer: %w", err)
	}
	defer closer()

	return launchStream(ctx, streamConfig{
		writer:      writer,
		stats:       newStats(),
		brange:      inputs.Range,
		cursor:      startCursor,
		endpoint:    endpoint,
		handleForks: viper.GetBool("global-handle-forks"),
		transforms:  transforms,
	},
		func() proto.Message {
			return &pbcodec.Block{}
		},
		func(message proto.Message) sf.BlockRef {
			b := message.(*pbcodec.Block)
			return sf.NewBlockRef(b.Header.Hash.String(), b.Header.Height)
		})
}
