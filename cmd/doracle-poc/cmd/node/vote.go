package node

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/client"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func CmdVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [vote_target_address]",
		Short: "vote registering oracle node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli, txBuilder, err := generateGrpcClientAndTxBuilder(cmd)
			if err != nil {
				return err
			}

			chainID, err := cmd.Flags().GetString("chain-id")
			if err != nil {
				return err
			}

			voteTargetAddress := args[0]
			err = verify(cmd, cli, voteTargetAddress)
			if err != nil {
				return err
			}

			mnemonic, err := cmd.Flags().GetString("mnemonic")
			if err != nil {
				return err
			}
			oracleAccPrivKey, err := panacea.GetPrivateKey(mnemonic)
			if err != nil {
				return err
			}
			oracleAddress, err := panacea.GetAddress(oracleAccPrivKey.PubKey())
			if err != nil {
				return err
			}

			oraclePrivateKey, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
			if err != nil {
				return err
			}

			encryptedData, err := generateEncryptedOraclePrivateKey(cli, oraclePrivateKey, voteTargetAddress)
			if err != nil {
				return err
			}

			unsignedVotingOracle := &oracletypes.UnsignedVotingOracle{
				Voter:              oracleAddress,
				VotingTarget:       voteTargetAddress,
				EncryptedOracleKey: encryptedData,
			}
			unsignedVotingOracleBz, err := unsignedVotingOracle.Marshal()
			if err != nil {
				return err
			}
			sign, err := secp256k1.PrivKeySecp256k1FromBytes(oraclePrivateKey).Sign(unsignedVotingOracleBz)
			if err != nil {
				return err
			}
			msg := &oracletypes.MsgVoteRegisteringOracle{
				VotingOracle: &oracletypes.VotingOracle{
					UnsignedVotingOracle: unsignedVotingOracle,
					Signature:            sign,
				},
			}

			txBytes, err := txBuilder.GenerateSignedTxBytes(oracleAccPrivKey, chainID, 300000, msg)
			if err != nil {
				return err
			}

			res, err := cli.Broadcast(txBytes)
			if err != nil {
				return err
			}
			log.Println(res)

			return nil
		},
	}

	cmd.Flags().String("grpcAddr", "", "blockChain grpc address")
	cmd.Flags().String("mnemonic", "", "Your mnemonic")
	cmd.Flags().String("chain-id", "", "Chain ID")
	cmd.Flags().String("signer-id", "", "signer ID")
	return cmd
}

func verify(cmd *cobra.Command, cli *client.GrpcClient, targetAddress string) error {
	signerID, err := cmd.Flags().GetString("signer-id")
	if err != nil {
		return err
	}
	oracle, err := cli.GetRegisterOracle(targetAddress)
	if err != nil {
		return err
	}

	if oracle.Status != "PENDING" {
		return fmt.Errorf("%s oracle status is not 'PENDING'", targetAddress)
	}

	report := oracle.RemoteReport
	data, err := sgx.VerifyRemoteReportAndGetData(report, signerID)
	if err != nil {
		return err
	}

	nodePubKeyBz := data[:btcec.PubKeyBytesLenCompressed]
	_, err = secp256k1.PubKeyFromBytes(nodePubKeyBz)
	if err != nil {
		return fmt.Errorf("failed parse to publicKey %w", err)
	}
	if !bytes.Equal(oracle.NodePublicKey, nodePubKeyBz) {
		return fmt.Errorf("is not same to nodePubKey. remoteReport(%s), chain(%s)", base64.StdEncoding.EncodeToString(nodePubKeyBz), base64.StdEncoding.EncodeToString(oracle.NodePublicKey))
	}

	log.Println("Verify success")
	return nil
}

func generateEncryptedOraclePrivateKey(cli *client.GrpcClient, oraclePrivateKeyBz []byte, voteTargetAddress string) ([]byte, error) {
	voteTargetOracle, err := cli.GetRegisterOracle(voteTargetAddress)
	if err != nil {
		return nil, err
	}
	nodePubKeyBz := voteTargetOracle.NodePublicKey
	if nodePubKeyBz == nil {
		return nil, fmt.Errorf("node publicKey is nil")
	}
	pubKey, err := secp256k1.PubKeyFromBytes(nodePubKeyBz)
	if err != nil {
		return nil, err
	}

	shareKeyBz := secp256k1.ShareKey(secp256k1.PrivKeyFromBytes(oraclePrivateKeyBz), pubKey)

	return secp256k1.EncryptWithAES256(shareKeyBz, nodePubKeyBz, nodePubKeyBz[:12], oraclePrivateKeyBz)
}
