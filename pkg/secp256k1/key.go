package secp256k1

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

func NewPrivKey() (*btcec.PrivateKey, error) {
	return btcec.NewPrivateKey(btcec.S256())
}

func PrivKeyFromBytes(bytes []byte) *btcec.PrivateKey {
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), bytes)
	return privKey
}

func PubKeyFromBytes(bytes []byte) (*btcec.PublicKey, error) {
	return btcec.ParsePubKey(bytes, btcec.S256())
}

func PrivKeySecp256k1FromBytes(bytes []byte) *secp256k1.PrivKey {
	return &secp256k1.PrivKey{
		Key: bytes,
	}
}

func PubKeySecp256k1FromBytes(bytes []byte) *secp256k1.PubKey {
	return &secp256k1.PubKey{
		Key: bytes,
	}
}

func ShareKey(priv *btcec.PrivateKey, pub *btcec.PublicKey) []byte {
	return btcec.GenerateSharedSecret(priv, pub)
}
