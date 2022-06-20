package client

import (
	"fmt"
	"github.com/medibloc/doracle-poc/pkg/client"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/spf13/cobra"
	"io/fs"
	"io/ioutil"
)

func CmdGenerateEncryptData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-encrypt-data [data]",
		Short: "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcAddr, err := cmd.Flags().GetString("grpcAddr")
			if err != nil {
				return err
			}

			conf := panacea.NewConfig()
			cli, err := client.NewGrpcClient(conf, grpcAddr)

			oraclePublicKeyBz, err := cli.GetOraclePublicKey()
			if err != nil {
				return err
			}
			oraclePublicKey, err := secp256k1.PubKeyFromBytes(oraclePublicKeyBz)
			if err != nil {
				return err
			}

			data := args[0]
			if data == "" {
				return fmt.Errorf("data is empty")
			}

			encryptedData, err := secp256k1.Encrypt(oraclePublicKey, []byte(data))
			if data == "" {
				return fmt.Errorf("data is empty")
			}
			outputPath, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(outputPath, encryptedData, fs.FileMode(0644))
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("grpcAddr", "", "blockChain grpc address")
	cmd.Flags().String("output", "encrypted_file", "encrypted file output path")

	return cmd
}
