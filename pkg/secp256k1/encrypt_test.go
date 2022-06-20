package secp256k1

import (
	"bytes"
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestSameEncryptData(t *testing.T) {
	privKey, err := NewPrivKey()
	require.NoError(t, err)

	data := []byte("test data")

	encrypted1, err := Encrypt(privKey.PubKey(), data)
	require.NoError(t, err)
	encrypted2, err := Encrypt(privKey.PubKey(), data)
	require.NoError(t, err)

	require.False(t, bytes.Equal(encrypted1, encrypted2))
}

func TestShareKey(t *testing.T) {
	privKey1, err := NewPrivKey()
	require.NoError(t, err)
	privKey2, err := NewPrivKey()
	require.NoError(t, err)

	shareKey1 := ShareKey(privKey1, privKey2.PubKey())
	shareKey2 := ShareKey(privKey2, privKey1.PubKey())

	require.True(t, bytes.Equal(shareKey1, shareKey2))
}

func TestEncryptedAES256(t *testing.T) {
	privKey1, err := NewPrivKey()
	require.NoError(t, err)
	privKey2, err := NewPrivKey()
	require.NoError(t, err)
	data := []byte("test data")

	shareKey1 := ShareKey(privKey1, privKey2.PubKey())
	shareKey2 := ShareKey(privKey2, privKey1.PubKey())

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)

	addition := privKey2.PubKey().SerializeCompressed()

	encryptData1, err := EncryptWithAES256(shareKey1, addition, nonce, data)
	require.NoError(t, err)
	encryptData2, err := EncryptWithAES256(shareKey2, addition, nonce, data)
	require.NoError(t, err)

	require.Equal(t, encryptData1, encryptData2)
}
