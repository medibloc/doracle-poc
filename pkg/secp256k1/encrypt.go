package secp256k1

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// Encrypt encrypts data using a secp256k1 public key (ECIES)
func Encrypt(pubKey *btcec.PublicKey, data []byte) ([]byte, error) {
	return btcec.Encrypt(pubKey, data)
}

// Decrypt decrypts data using a secp256k1 private key (ECIES)
func Decrypt(privKey *btcec.PrivateKey, data []byte) ([]byte, error) {
	return btcec.Decrypt(privKey, data)
}

func GeneratePrivateKeyFromMnemonic(mnemonic string, coinType, accountNumber, index uint32) (secp256k1.PrivKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	hdPath := hd.NewFundraiserParams(accountNumber, coinType, index).String()
	master, ch := hd.ComputeMastersFromSeed(bip39.NewSeed(mnemonic, ""))

	return hd.DerivePrivateKeyForPath(master, ch, hdPath)
}

func EncryptWithAES256(secretKey, additional, nonce, data []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	// prepare AES-256-GSM cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != aesGCM.NonceSize() {
		return nil, fmt.Errorf("nonce length must be %v", aesGCM.NonceSize())
	}
	// encrypt data with second key
	ciphertext := aesGCM.Seal(nonce, nonce, data, additional)
	return ciphertext, nil
}

func DecryptWithAES256(secretKey, additional []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, pureCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// decrypt ciphertext with second key
	plaintext, err := aesgcm.Open(nil, nonce, pureCiphertext, additional)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
