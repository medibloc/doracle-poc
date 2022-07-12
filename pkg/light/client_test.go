package light

import (
	"context"
	"encoding/hex"
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	hash, err := hex.DecodeString("0B07B2C2836E7E249516208235EF2700B37A202FA1B84F3DFBD28A7A6703352B")
	if err != nil {
		return nil, err
	}
	trustOptions := light.TrustOptions{
		Period: 4 * time.Hour,
		Height: 2,
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

func TestLightClient(t *testing.T) {
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

func TestProof(t *testing.T) {
	rpcClient, err := httprpc.New("http://localhost:26657", "/websocket")
	require.NoError(t, err)

	ctx := context.Background()
	lc, err := makeLightClient(ctx)
	require.NoError(t, err)
	trustedBlock, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)
	fmt.Println(trustedBlock)

	acc, err := panacea.GetAccount("panacea19lnpp2w657r6petyqyydu5mk0xpnxryhzqjuza")
	require.NoError(t, err)
	key := authtypes.AddressStoreKey(acc)
	option := rpcclient.ABCIQueryOptions{
		Prove:  true,
		Height: trustedBlock.Height,
	}

	result, err := rpcClient.ABCIQueryWithOptions(ctx, "/store/acc/key", key, option)
	require.NoError(t, err)

	if !result.Response.IsOK() {
		panic(result.Response)
	}

	/*fmt.Println(rpcClient.ConsensusState(ctx))

	merkleProof, err := types.ConvertProofs(result.Response.ProofOps)
	require.NoError(t, err)

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(trustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath("acc", string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	require.NoError(t, err)*/

}
