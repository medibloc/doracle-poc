package client

import "github.com/spf13/cobra"

func CmdClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client [subCommand]",
		Short: "client",
	}

	cmd.AddCommand(CmdGenerateEncryptData())

	return cmd
}
