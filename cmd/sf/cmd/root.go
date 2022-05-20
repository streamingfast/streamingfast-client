package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const usage = `
Connects to StreamingFast endpoint using the STREAMINGFAST_API_KEY from
environment variables and stream back blocks filterted using the <filter>
argument within the <start_block> and <end_block> if they are specified.

If STREAMINGFAST_API_KEY environment is not set, only unauthenticated network
will be connectable to, authenticated network will refuse the connection.

Parameters:
  <start_block>   Optional block number where to start streaming blocks from,
                  Can be positive (an absolute reference to a block), or
                  negative (a number of blocks from the tip of the chain).

  <end_block>     Optional block number end block boundary after which (inclusively)
				  the stream of blocks will stop If not specified, the stream
				  will stop when the Ethereum network stops: never.
Examples:


  # Look at ALL blocks in a given range on ETH mainnet
  $ sf eth 100000 100010

  # Stream blocks in a given range on BSC with logs that match a given address and event signature (topic0)
  #   (transactions that do not match are filtered out of the "transactionTraces" array in the response)
  # sf eth --bsc --log-filter-addresses='0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73' --log-filter-event-sigs='0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9' 6500000 6810800

  # Continue where you left off, start from the last known cursor, get all fork notifications (UNDO, IRREVERSIBLE), stream forever
  $ sf eth --handle-forks --start-cursor "10928019832019283019283" "to in ['0x7a250d5630b4cf539739df2c5dacb4c659f2488d']"

  # Stream blocks from the last 5 on NEAR testnet
  $ sf near --testnet -- -5

  # Look at ALL blocks in a given range on Polygon Chain
  $ sf eth --polygon 100000 100010
`

// RootCmd represents the eosc command
var RootCmd = &cobra.Command{
	Use:   "sf",
	Short: "Streaming Fast command-line client",
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().Bool("handle-forks", false, "Request notifications type STEP_UNDO when a block was forked out, and STEP_IRREVERSIBLE after a block has seen enough confirmations (200)")
	RootCmd.PersistentFlags().BoolP("insecure", "s", false, "Use TLS for the connection with the remote server but skips SSL certificate verification, this is an insecure setup")
	RootCmd.PersistentFlags().BoolP("plaintext", "p", false, "Use plain text for the connection with the remote server, this is an insecure setup and any middleman is able to see the traffic")
	RootCmd.PersistentFlags().BoolP("skip-auth", "a", false, "Skips the authentication")
	RootCmd.PersistentFlags().StringP("output", "o", "-", "When set, write each block as one JSON line in the specified file, value '-' writes to standard output otherwise to a file, {range} is replaced by block range in this case")
	RootCmd.PersistentFlags().String("start-cursor", "", "Last cursor used to continue where you left off")
	RootCmd.PersistentFlags().String("auth-endpoint", "", "Use an alternative authentication endpoint for retrieving access tokens.")
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
