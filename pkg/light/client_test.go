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
	rpcClient, err := httprpc.New(rpcAddr, "/websocket")
	require.NoError(t, err)

	ctx := context.Background()
	lc, err := makeLightClient(ctx)
	require.NoError(t, err)
	trustedBlock, err := lc.Update(ctx, time.Now())
	require.NoError(t, err)
	fmt.Println(trustedBlock)

	acc, err := panacea.GetAccount("panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3")
	require.NoError(t, err)
	key := authtypes.AddressStoreKey(acc)
	option := rpcclient.ABCIQueryOptions{
		Prove:  true,
		Height: trustedBlock.Height - 1,
	}

	result, err := rpcClient.ABCIQueryWithOptions(ctx, fmt.Sprintf("/store/%s/key", authtypes.StoreKey), key, option)
	require.NoError(t, err)

	any := &types2.Any{}
	err = any.Unmarshal(result.Response.Value)
	require.NoError(t, err)

	var getAcc authtypes.AccountI
	err = panacea.NewConfig().InterfaceRegistry.UnpackAny(any, &getAcc)
	require.NoError(t, err)

	proofOps := result.Response.ProofOps
	fmt.Println(proofOps)

	if !result.Response.IsOK() {
		panic(result.Response)
	}

	merkleProof, err := types.ConvertProofs(proofOps)
	require.NoError(t, err)

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(trustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath("acc", string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	require.NoError(t, err)
}
