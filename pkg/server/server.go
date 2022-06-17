package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/gorilla/mux"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	oraclePrivKey *btcec.PrivateKey
}

func NewServer(oraclePrivKey *btcec.PrivateKey) *Server {
	return &Server{
		oraclePrivKey: oraclePrivKey,
	}
}

type ShutdownFunc func()

func (s *Server) ListenAndServe(listenAddr string) ShutdownFunc {
	router := mux.NewRouter()
	router.HandleFunc("/handshake", s.HandleHandshake).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Infof("Listening %s...", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Errorf("http server was shutted down: %v", err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		srv.Shutdown(ctx)
	}
}

func (s *Server) HandleHandshake(w http.ResponseWriter, r *http.Request) {
	respBodyBytes, err := s.handleHandshake(r)
	if err != nil {
		log.Errorf("failed to handle join: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(respBodyBytes); err != nil {
		log.Printf("failed to write response body: %v", err)
		return
	}
}

func (s *Server) handleHandshake(r *http.Request) ([]byte, error) {
	reqBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	var reqBody HandshakeRequestBody
	if err := json.Unmarshal(reqBodyBytes, &reqBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	reportBytes, err := base64.StdEncoding.DecodeString(reqBody.ReportBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ReportBase64: %w", err)
	}
	peerNodePubKeyBytes, err := base64.StdEncoding.DecodeString(reqBody.NodePublicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode NodePublicKeyBase64: %w", err)
	}

	pubKeyHash := sha256.Sum256(peerNodePubKeyBytes)
	if err := sgx.VerifyRemoteReport(reportBytes, pubKeyHash[:]); err != nil {
		return nil, fmt.Errorf("failed to verify SGX remote report: %w", err)
	}

	peerNodePubKey, err := secp256k1.PubKeyFromBytes(peerNodePubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the peer node public key bytes: %w", err)
	}

	encryptedOraclePrivKey, err := secp256k1.Encrypt(peerNodePubKey, s.oraclePrivKey.Serialize())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt oracle private key: %w", err)
	}

	respBody := HandshakeResponseBody{
		EncryptedOraclePrivateKeyBase64: base64.StdEncoding.EncodeToString(encryptedOraclePrivKey),
	}
	respBodyBytes, err := json.Marshal(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body: %w", err)
	}
	return respBodyBytes, nil
}
