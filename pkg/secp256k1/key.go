package secp256k1

import (
	"github.com/btcsuite/btcd/btcec"
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
