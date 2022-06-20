package node

import (
	"github.com/medibloc/doracle-poc/pkg/client"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/medibloc/doracle-poc/pkg/panacea/tx"
	"github.com/spf13/cobra"
)

func CmdNode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node [subCommand]",
		Short: "node",
	}

	cmd.AddCommand(CmdRegister())
	cmd.AddCommand(CmdInitOracleKey())
	cmd.AddCommand(CmdVote())
	cmd.AddCommand(CmdGetOracleKey())
	cmd.AddCommand(CmdReadEncryptedFile())

	return cmd
}

func generateGrpcClientAndTxBuilder(cmd *cobra.Command) (*client.GrpcClient, *tx.TxBuilder, error) {
	grpcAddr, err := cmd.Flags().GetString("grpcAddr")
	if err != nil {
		return nil, nil, err
	}

	conf := panacea.NewConfig()
	cli, err := client.NewGrpcClient(conf, grpcAddr)
	if err != nil {
		return nil, nil, err
	}

	return cli, tx.NewTxBuilder(conf, cli), nil

}
