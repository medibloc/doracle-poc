package node

import (
	"fmt"
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdGetOracleKey() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-oracle-key",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, _, err := generateGrpcClientAndTxBuilder(cmd)

			mnemonic, err := cmd.Flags().GetString("mnemonic")
			if err != nil {
				return err
			}
			privateKey, err := panacea.GetPrivateKey(mnemonic)
			if err != nil {
				return err
			}
			address, err := panacea.GetAddress(privateKey.PubKey())
			if err != nil {
				return err
			}

			oracle, err := cli.GetRegisterOracle(address)
			if err != nil {
				return err
			}
			if oracle.Status != oracletypes.REGISTER {
				return fmt.Errorf("not registed oracle. current status: %v", oracle.Status)
			}

			encryptedKey := oracle.EncryptedOracleKey

			nodePrivateKeyBz, err := sgx.UnsealFromFile(mode.NodeKeyFilePath)
			if err != nil {
				return err
			}
			nodePrivKey := secp256k1.PrivKeyFromBytes(nodePrivateKeyBz)
			if err != nil {
				return err
			}
			oraclePublicKeyBz, err := cli.GetOraclePublicKey()
			if err != nil {
				return err
			}
			oraclePublicKey, err := secp256k1.PubKeyFromBytes(oraclePublicKeyBz)
			if err != nil {
				return err
			}

			shareKeyBz := secp256k1.ShareKey(nodePrivKey, oraclePublicKey)

			oraclePrivateKeyBz, err := secp256k1.DecryptWithAES256(shareKeyBz, nodePrivKey.PubKey().SerializeCompressed(), encryptedKey)
			if err != nil {
				return err
			}

			err = sgx.SealToFile(oraclePrivateKeyBz, mode.OracleKeyFilePath)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String("grpcAddr", "", "blockChain grpc address")
	cmd.Flags().String("mnemonic", "", "Your mnemonic")
	cmd.Flags().String("chain-id", "", "Chain ID")
	return cmd
}
