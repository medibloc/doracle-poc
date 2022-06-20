package node

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/client"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/medibloc/doracle-poc/pkg/panacea/tx"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func CmdRegister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-node",
		Short: "Register oracle node",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, txBuilder, err := generateGrpcClientAndTxBuilder(cmd)

			mnemonic, err := cmd.Flags().GetString("mnemonic")
			if err != nil {
				return err
			}

			chainID, err := cmd.Flags().GetString("chain-id")
			if err != nil {
				return err
			}

			err = registerOracle(cli, txBuilder, mnemonic, chainID)
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

func registerOracle(
	client *client.GrpcClient,
	builder *tx.TxBuilder,
	mnemonic, chainID string,
) error {
	privateKey, err := panacea.GetPrivateKey(mnemonic)
	if err != nil {
		return err
	}
	address, err := panacea.GetAddress(privateKey.PubKey())
	if err != nil {
		return err
	}

	msg, err := generateRegisterOracleMsg(address)
	if err != nil {
		return err
	}

	txBytes, err := builder.GenerateSignedTxBytes(privateKey, chainID, 300000, msg)
	if err != nil {
		return err
	}

	res, err := client.Broadcast(txBytes)
	log.Println(res)

	return nil
}

func generateRegisterOracleMsg(address string) (*oracletypes.MsgRegisterOracle, error) {
	nodePrivKey, err := secp256k1.NewPrivKey()
	if err != nil {
		return nil, err
	}
	err = sgx.SealToFile(nodePrivKey.Serialize(), mode.NodeKeyFilePath)
	if err != nil {
		return nil, err
	}
	nodePubKeyBytes := nodePrivKey.PubKey().SerializeCompressed()
	report, err := sgx.GenerateRemoteReport(nodePubKeyBytes[:])
	if err != nil {
		return nil, err
	}
	oracleMsg := &oracletypes.MsgRegisterOracle{
		OracleDetail: &oracletypes.Oracle{
			Address:       address,
			Endpoint:      "https://my-oracle.org",
			RemoteReport:  report,
			NodePublicKey: nodePubKeyBytes[:],
		},
	}
	return oracleMsg, nil
}

func CmdInitOracleKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-oracle-key",
		Short: "generate oracle privateKey and set oracle publicKey to blockchain",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := generateOracleKey()
			if err != nil {
				return err
			}

			err = setOraclePublicKeyToPanacea(cmd)
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

func generateOracleKey() error {
	return mode.Init()
}

func setOraclePublicKeyToPanacea(cmd *cobra.Command) error {
	cli, txBuilder, err := generateGrpcClientAndTxBuilder(cmd)

	mnemonic, err := cmd.Flags().GetString("mnemonic")
	if err != nil {
		return err
	}
	oraclePrivateKey, err := panacea.GetPrivateKey(mnemonic)
	if err != nil {
		return err
	}

	chainID, err := cmd.Flags().GetString("chain-id")
	if err != nil {
		return err
	}

	msg, err := generateOraclePublicKey(err, oraclePrivateKey)
	if err != nil {
		return err
	}

	txBytes, err := txBuilder.GenerateSignedTxBytes(oraclePrivateKey, chainID, 300000, msg)
	if err != nil {
		return err
	}

	res, err := cli.Broadcast(txBytes)
	log.Println(res)

	return nil
}

func generateOraclePublicKey(err error, oraclePrivateKey cryptotypes.PrivKey) (*oracletypes.MsgOraclePublicKey, error) {
	oracleKey, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
	if err != nil {
		return nil, err
	}

	oraclePublicKey := secp256k1.PrivKeyFromBytes(oracleKey).PubKey()
	oracleAddress, err := panacea.GetAddress(oraclePrivateKey.PubKey())
	if err != nil {
		return nil, err
	}

	msg := &oracletypes.MsgOraclePublicKey{
		OraclePublicKey: oraclePublicKey.SerializeCompressed(),
		OracleAddress:   oracleAddress,
	}
	return msg, nil
}
