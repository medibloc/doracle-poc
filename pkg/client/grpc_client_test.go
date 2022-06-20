package client_test

import (
	"github.com/medibloc/doracle-poc/pkg/client"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/medibloc/doracle-poc/pkg/panacea/tx"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSend(t *testing.T) {
	mnemonic := "effort kite tell stuff beauty chest bag noise verify salute laundry eyebrow rally main plunge dwarf venue chief sing vicious lend napkin raccoon airport"
	privateKey, err := panacea.GetPrivateKey(mnemonic)
	require.NoError(t, err)

	conf := panacea.NewConfig()
	cli, err := client.NewGrpcClient(conf, "http://localhost:9090")
	require.NoError(t, err)

	txBuilder := tx.NewTxBuilder(conf, cli)

	/*msg := &banktypes.MsgSend{
		FromAddress: "panacea13r2mlszg22748s3s5x4z2rxaya5hdzzh992yzq",
		ToAddress:   "panacea1xwudeh8pthyv2h3nygy6lesnarw39z7vya34uu",
		Amount:      sdk.NewCoins(sdk.NewCoin("umed", sdk.NewInt(1))),
	}*/
	msg := &oracletypes.MsgRegisterOracle{
		OracleDetail: &oracletypes.Oracle{
			Address:      "panacea13r2mlszg22748s3s5x4z2rxaya5hdzzh992yzq",
			Endpoint:     "https://my-oracle.org",
			RemoteReport: nil,
		},
	}

	txBytes, err := txBuilder.GenerateSignedTxBytes(privateKey, "gyuguen-1", 200000, msg)
	require.NoError(t, err)

	res, err := cli.Broadcast(txBytes)
	require.NoError(t, err)
	log.Println(res)
}
