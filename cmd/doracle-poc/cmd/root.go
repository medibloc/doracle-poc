package cmd

import (
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/cmd/client"
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/cmd/node"
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/cmd/subscribe"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "doracle-poc",
		Short: "doracle-poc daemon",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(node.CmdNode())
	rootCmd.AddCommand(client.CmdClient())
	rootCmd.AddCommand(subscribe.CmdSubscribe())
}
