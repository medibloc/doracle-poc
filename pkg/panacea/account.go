package panacea

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
)

const (
	hrp      = "panacea"
	coinType = 371
)

func GetPrivateKey(mnemonic string) (cryptotypes.PrivKey, error) {
	key, err := secp256k1.GeneratePrivateKeyFromMnemonic(mnemonic, coinType, 0, 0)
	if err != nil {
		return nil, err
	}

	return secp256k1.PrivKeySecp256k1FromBytes(key.Bytes()), nil
}

func GetAddress(publicKey cryptotypes.PubKey) (string, error) {
	return bech32.ConvertAndEncode(hrp, publicKey.Address().Bytes())
}

func GetAccount(address string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(address, hrp)
}
