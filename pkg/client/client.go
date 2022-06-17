package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/server"
	"github.com/medibloc/doracle-poc/pkg/sgx"
)

func CallHandshake(peerAddr string, nodePrivKey *btcec.PrivateKey) (*btcec.PrivateKey, error) {
	nodePubKeyBytes := nodePrivKey.PubKey().SerializeCompressed()
	pubKeyHash := sha256.Sum256(nodePubKeyBytes)
	report, err := sgx.GenerateRemotePeport(pubKeyHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate SGX remote report: %w", err)
	}

	reqBody := server.HandshakeRequestBody{
		ReportBase64:        base64.StdEncoding.EncodeToString(report),
		NodePublicKeyBase64: base64.StdEncoding.EncodeToString(nodePubKeyBytes),
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/handshake", peerAddr), "application/json", bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to call handshake: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the handshake request wasn't accepted. status:%d", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var respBody server.HandshakeResponseBody
	if err := json.Unmarshal(respBodyBytes, &respBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	encryptedOraclePrivKey, err := base64.StdEncoding.DecodeString(respBody.EncryptedOraclePrivateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EncryptedOraclePrivateKeyBase64: %w", err)
	}

	oraclePrivKeyBytes, err := secp256k1.Decrypt(nodePrivKey, encryptedOraclePrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt encryptedOraclePrivKey: %w", err)
	}

	return secp256k1.PrivKeyFromBytes(oraclePrivKeyBytes), nil
}
