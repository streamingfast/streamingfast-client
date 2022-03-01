package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sf "github.com/streamingfast/streamingfast-client"
	pbcodec "github.com/streamingfast/streamingfast-client/pb/sf/solana/codec/v1"
	"google.golang.org/protobuf/proto"
)

var solSfCmd = &cobra.Command{
	Use:   "sol [flags] [<start_block>] [<end_block>]",
	Short: `StreamingFast Solana client`,
	Long:  usage,
	RunE:  solSfCmdE,
}

func init() {
	RootCmd.AddCommand(solSfCmd)

	// Transforms
	solSfCmd.Flags().String("program-filter", "", "Specifcy a comma delimited base58 program addresses to filter a block for given programs")

	defaultSolEndpoint := "mainnet.sol.streamingfast.io:443"
	if e := os.Getenv("STREAMINGFAST_ENDPOINT"); e != "" {
		defaultSolEndpoint = e
	}
	solSfCmd.Flags().StringP("endpoint", "e", defaultSolEndpoint, "The endpoint to connect the stream of blocks (default value set by STREAMINGFAST_ENDPOINT env var, can be overriden by network-specific flags like --testnet)")
	solSfCmd.Flags().Bool("testnet", false, "When set, will switch default endpoint testnet")

}

func solSfCmdE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	startCursor := viper.GetString("global-start-cursor")
	endpoint := viper.GetString("sol-cmd-endpoint")
	testnet := viper.GetBool("sol-cmd-testnet")
	outputFlag := viper.GetString("global-output")

	inputs, err := checkArgs(startCursor, args)
	if err != nil {
		return err
	}

	if testnet {
		endpoint = "testnet.sol.streamingfast.io:443"
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
	},
		func() proto.Message {
			return &pbcodec.Block{}
		},
		func(message proto.Message) sf.BlockRef {
			return message.(*pbcodec.Block).AsRef()
		})
}
