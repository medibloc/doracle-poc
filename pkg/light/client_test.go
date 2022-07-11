package light

import (
	"context"
	"encoding/hex"
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	"github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	httprpc "github.com/tendermint/tendermint/rpc/client/http"
	dbm "github.com/tendermint/tm-db"
	"testing"
	"time"
)

func makeLightClient(ctx context.Context) (*light.Client, error) {
	hash, err := hex.DecodeString("746BEE92DB872B2DC4699D28C95CC9F99FF1A2E0B191590899D8D6EFC7A08C7D")
	if err != nil {
		return nil, err
	}
	trustOptions := light.TrustOptions{
		Period: 4 * time.Hour,
		Height: 4001,
		Hash:   hash,
	}
	pv, err := http.New("gyuguen-1", "http://localhost:26657")
	if err != nil {
		return nil, err
	}

	pvs := []provider.Provider{pv}
	store := dbs.New(dbm.NewMemDB(), "gyuguen-1")
	lc, err := light.NewClient(
		ctx,
		"gyuguen-1",
		trustOptions,
		pv,
		pvs,
		store,
		light.SequentialVerification(),
		light.Logger(log.TestingLogger()),
	)
	return lc, nil
}

func Test(t *testing.T) {
	ctx := context.Background()
	lc, err := makeLightClient(ctx)
	require.NoError(t, err)

	tb, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)
	fmt.Println(tb)
	fmt.Println(lc.FirstTrustedHeight())
	fmt.Println(lc.LastTrustedHeight())
	fmt.Println(lc.TrustedLightBlock(tb.Height))
}

func TestSubscribe(t *testing.T) {
	rpcClient, err := httprpc.New("http://localhost:26657", "/websocket")
	require.NoError(t, err)

	ctx := context.Background()
	lc, err := makeLightClient(ctx)
	require.NoError(t, err)
	trustedBlock, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)

	beforeTrustedBlock, err := lc.VerifyLightBlockAtHeight(ctx, trustedBlock.Height-1, time.Now())
	require.NoError(t, err)
	fmt.Println(beforeTrustedBlock)

	acc, err := panacea.GetAccount("panacea17exeeteqq7g82g8nqvmm8kfrs4qm66ulfgmj6t")
	require.NoError(t, err)
	key := authtypes.AddressStoreKey(acc)
	option := rpcclient.ABCIQueryOptions{
		Prove:  true,
		Height: beforeTrustedBlock.Height,
	}

	consensusStatus, err := rpcClient.ConsensusState(ctx)
	fmt.Println("consensus:", string(consensusStatus.RoundState))

	result, err := rpcClient.ABCIQueryWithOptions(ctx, "/store/acc/key", key, option)
	require.NoError(t, err)

	if !result.Response.IsOK() {
		panic(result.Response)
	}

	merkleProof, err := types.ConvertProofs(result.Response.ProofOps)
	require.NoError(t, err)
	fmt.Println(merkleProof)

	merkleProof.VerifyMembership()

}
