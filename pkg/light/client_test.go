package light

import (
	"context"
	"encoding/hex"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
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

const (
	chainID = "panacea-3"
	rpcAddr = "https://rpc.gopanacea.org:443"
)

var (
	ctx = context.Background()
)

func makeLightClient(ctx context.Context) (*light.Client, error) {
	hash, err := hex.DecodeString("3531F0F323110AA7831775417B9211348E16A29A07FBFD46018936625E4E5492")
	if err != nil {
		return nil, err
	}
	trustOptions := light.TrustOptions{
		Period: 365 * 24 * time.Hour,
		Height: 99,
		Hash:   hash,
	}
	pv, err := http.New(chainID, rpcAddr)
	if err != nil {
		return nil, err
	}

	pvs := []provider.Provider{pv}
	store := dbs.New(dbm.NewMemDB(), chainID)
	lc, err := light.NewClient(
		ctx,
		chainID,
		trustOptions,
		pv,
		pvs,
		store,
		light.SkippingVerification(light.DefaultTrustLevel),
		light.Logger(log.TestingLogger()),
	)
	return lc, nil
}

func TestLightClientSyncLastBlock(t *testing.T) {
	lc, err := makeLightClient(ctx)
	require.NoError(t, err)

	tb, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)
	fmt.Println(tb)
}

func TestLightClient(t *testing.T) {
	lc, err := makeLightClient(ctx)
	require.NoError(t, err)

	fmt.Println(lc.VerifyLightBlockAtHeight(ctx, 6167926, time.Now()))

}

func TestProof(t *testing.T) {
	rpcClient, err := httprpc.New(rpcAddr, "/websocket")
	require.NoError(t, err)

	lc, err := makeLightClient(ctx)
	require.NoError(t, err)
	trustedBlock, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)
	fmt.Println(trustedBlock)

	// MediblocLimit-1 address
	acc, err := panacea.GetAccount("panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3")
	require.NoError(t, err)
	key := authtypes.AddressStoreKey(acc)
	option := rpcclient.ABCIQueryOptions{
		Prove:  true,
		Height: trustedBlock.Height - 1,
	}

	result, err := rpcClient.ABCIQueryWithOptions(ctx, fmt.Sprintf("/store/%s/key", authtypes.StoreKey), key, option)
	require.NoError(t, err)

	resultAccBz := result.Response.Value

	resultAccAny := &types2.Any{}
	err = resultAccAny.Unmarshal(resultAccBz)
	require.NoError(t, err)

	var resultAcc authtypes.AccountI
	err = panacea.NewConfig().InterfaceRegistry.UnpackAny(resultAccAny, &resultAcc)
	require.NoError(t, err)

	resultAddress, err := panacea.GetAddress(resultAcc.GetPubKey())
	require.NoError(t, err)
	fmt.Println("address: ", resultAddress)

	proofOps := result.Response.ProofOps
	fmt.Println("proofOps: ", proofOps)

	if !result.Response.IsOK() {
		panic(result.Response)
	}

	merkleProof, err := types.ConvertProofs(proofOps)
	require.NoError(t, err)

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(trustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath(authtypes.StoreKey, string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	require.NoError(t, err)
}
