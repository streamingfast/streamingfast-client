package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const usage = `
Connects to StreamingFast endpoint using the STREAMINGFAST_API_KEY from
environment variables and stream back blocks filterted using the <filter>
argument within the <start_block> and <end_block> if they are specified.

If STREAMINGFAST_API_KEY environment is not set, only unauthenticated network
will be connectable to, authenticated network will refuse the connection.

Parameters:
  <filter>        Optional A valid CEL filter expression for the Ethereum network, only
                  transactions matching the filter will be returned to you.

  <start_block>   Optional block number where to start streaming blocks from,
                  Can be positive (an absolute reference to a block), or
                  negative (a number of blocks from the tip of the chain).

  <end_block>     Optional block number end block boundary after which (inclusively)
				  the stream of blocks will stop If not specified, the stream
				  will stop when the Ethereum network stops: never.
Examples:
  # Watch all calls to the UniswapV2 Router, for a single block and close
  $ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" 11700000 11700001

  # Watch all calls to the UniswapV2 Router, include the last 100 blocks, and stream forever
  $ sf "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']" -100

  # Continue where you left off, start from the last known cursor, get all fork notifications (UNDO, IRREVERSIBLE), stream forever
  $ sf --handle-forks --start-cursor "10928019832019283019283" "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']"

  # Look at ALL blocks in a given range on Binance Smart Chain (BSC)
  $ sf --bsc "true" 100000 100002

  # Look at ALL blocks in a given range on Polygon Chain
  $ sf --polygon "true" 100000 100002

  # Look at ALL blocks in a given range on Huobi ECO Chain
  $ sf --heco "true" 100000 100002

  # Look at recent blocks and stream forever on Fantom Opera Mainnet
  $ sf --fantom "true" -5

  # Look at recent blocks and stream forever on xDai Chain
  $ sf --xdai "true" -5
`

// RootCmd represents the eosc command
var RootCmd = &cobra.Command{
	Use:   "sf",
	Short: "Streaming Fast command-line client",
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP("endpoint", "e", "api.streamingfast.io:443", "The endpoint to connect the stream of blocks to")
	RootCmd.PersistentFlags().Bool("handle-forks", false, "Request notifications type STEP_UNDO when a block was forked out, and STEP_IRREVERSIBLE after a block has seen enough confirmations (200)")
	RootCmd.PersistentFlags().BoolP("insecure", "s", false, "Enables Insecure connection, When set, skips certification verification")
	RootCmd.PersistentFlags().BoolP("skip-auth", "a", false, "Skips the authentication")
	RootCmd.PersistentFlags().StringP("output", "o", "-", "When set, write each block as one JSON line in the specified file, value '-' writes to standard output otherwise to a file, {range} is replaced by block range in this case")
	RootCmd.PersistentFlags().String("start-cursor", "", "Last cursor used to continue where you left off")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Printf("exiting with error: %s", err)
		os.Exit(1)
	}

}

func initConfig() {
	viper.SetEnvPrefix("SF")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	recurseViperCommands(RootCmd, nil)
}

func recurseViperCommands(root *cobra.Command, segments []string) {
	// Stolen from: github.com/abourget/viperbind
	var segmentPrefix string
	if len(segments) > 0 {
		segmentPrefix = strings.Join(segments, "-") + "-"
	}

	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		newVar := segmentPrefix + "global-" + f.Name
		viper.BindPFlag(newVar, f)
	})
	root.Flags().VisitAll(func(f *pflag.Flag) {
		newVar := segmentPrefix + "cmd-" + f.Name
		viper.BindPFlag(newVar, f)
	})

	for _, cmd := range root.Commands() {
		recurseViperCommands(cmd, append(segments, cmd.Name()))
	}
}
