package tx

import (
	"fmt"
	tx2 "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/medibloc/doracle-poc/pkg/client"
	"github.com/medibloc/doracle-poc/pkg/panacea"
)

type TxBuilder struct {
	client     *client.GrpcClient
	marshaller *codec.ProtoCodec
}

func NewTxBuilder(conf *panacea.Config, client *client.GrpcClient) *TxBuilder {
	return &TxBuilder{
		client:     client,
		marshaller: conf.Marshaller,
	}
}

func (t TxBuilder) GenerateSignedTxBytes(
	privateKey cryptotypes.PrivKey,
	chainID string,
	gasLimit uint64,
	msg ...sdk.Msg) ([]byte, error) {

	txConfig := authtx.NewTxConfig(t.marshaller, []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})
	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetGasLimit(gasLimit)
	err := txBuilder.SetMsgs(msg...)
	if err != nil {
		return nil, err
	}

	signerAddress, err := panacea.GetAddress(privateKey.PubKey())
	if err != nil {
		return nil, err
	}

	signerAccount, err := t.client.GetAccount(signerAddress)
	if err != nil {
		return nil, err
	}

	sigV2 := signing.SignatureV2{
		PubKey: privateKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: signerAccount.GetSequence(),
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: signerAccount.GetAccountNumber(),
		Sequence:      signerAccount.GetSequence(),
	}
	sigV2, err = tx2.SignWithPrivKey(
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder,
		privateKey,
		txConfig,
		signerAccount.GetSequence(),
	)
	if err != nil {
		return nil, err
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	bz, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	fmt.Println(string(bz))

	return txConfig.TxEncoder()(txBuilder.GetTx())
}
