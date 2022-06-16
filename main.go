package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"github.com/gorilla/mux"
)

const keyFilePath = "/data/oracle-key.sealed"

var sampleReportData = []byte("hello, man")

func main() {
	pListenAddr := flag.String("laddr", "0.0.0.0:8080", "http listen addr")
	pInit := flag.Bool("init", false, "run doracle with the init mode")
	pPeer := flag.String("peer", "", "a peer addr (do not use with -init)")
	flag.Parse()

	if *pInit && *pPeer != "" {
		log.Fatal("do not use -peer with -init")
	}

	if *pInit {
		oracleKey, err := generateOracleKey()
		if err != nil {
			log.Fatalf("failed to generate oracle key: %v", err)
		}

		if err := saveOracleKey(oracleKey); err != nil {
			log.Fatalf("failed to save oracle key: %v", err)
		}
	}

	nodeKey, err := generateNodeKey()
	if err != nil {
		log.Fatalf("failed to generate node key: %v", err)
	}

	if *pPeer != "" {
		oracleKey, err := tryJoin(*pPeer, nodeKey)
		if err != nil {
			log.Fatalf("failed to join: %v", err)
		}

		if err := saveOracleKey(oracleKey); err != nil {
			log.Fatalf("failed to save oracle key: %v", err)
		}
	}

	oraclePrivKey, err := loadOraclePrivKey()
	if err != nil {
		log.Fatalf("failed to load oracle key: %v", err)
	}

	myRouter := &MyRouter{
		oraclePrivKey: oraclePrivKey,
	}

	router := mux.NewRouter()
	router.HandleFunc("/join", myRouter.HandleJoin).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         *pListenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening %s...", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func generateOracleKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func saveOracleKey(key *rsa.PrivateKey) error {
	oraclePrivKeyBz := x509.MarshalPKCS1PrivateKey(key)
	sealed, err := ecrypto.SealWithProductKey(oraclePrivKeyBz, nil)
	if err != nil {
		return fmt.Errorf("failed to seal oracle key: %w", err)
	}

	if err := ioutil.WriteFile(keyFilePath, sealed, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", keyFilePath, err)
	}
	log.Printf("%s is written successfully", keyFilePath)

	return nil
}

func loadOraclePrivKey() (*rsa.PrivateKey, error) {
	sealed, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", keyFilePath, err)
	}

	unsealed, err := ecrypto.Unseal(sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle key: %w", err)
	}

	privKey, err := x509.ParsePKCS1PrivateKey(unsealed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse oracle key: %w", err)
	}

	return privKey, nil
}

func generateNodeKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func tryJoin(peerAddr string, nodePrivKey *rsa.PrivateKey) (*rsa.PrivateKey, error) {
	report, err := enclave.GetRemoteReport(sampleReportData)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote report: %w", err)
	}

	reqBody := JoinRequestBody{
		ReportBase64:        base64.StdEncoding.EncodeToString(report),
		NodePublicKeyBase64: base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&nodePrivKey.PublicKey)),
	}
	reqBodyBz, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/join", peerAddr), "application/json", bytes.NewReader(reqBodyBz))
	if err != nil {
		return nil, fmt.Errorf("failed to call join: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the join request wasn't accepted. status:%d", resp.StatusCode)
	}

	respBodyBz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var respBody JoinResponseBody
	if err := json.Unmarshal(respBodyBz, &respBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	encryptedOracleKey, err := base64.StdEncoding.DecodeString(respBody.EncryptedOracleKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EncryptedOracleKeyBase64: %w", err)
	}

	oracleKeyBz, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, nodePrivKey, encryptedOracleKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt EncryptedOracleKey: %w", err)
	}

	oracleKey, err := x509.ParsePKCS1PrivateKey(oracleKeyBz)
	if err != nil {
		return nil, fmt.Errorf("failed to parse oracle key: %w", err)
	}

	return oracleKey, nil
}

type MyRouter struct {
	oraclePrivKey *rsa.PrivateKey
}

type JoinRequestBody struct {
	ReportBase64        string `json:"report_base64"`
	NodePublicKeyBase64 string `json:"node_public_key_base64"`
}

type JoinResponseBody struct {
	EncryptedOracleKeyBase64 string `json:"encrypted_oracle_key_base64"`
}

func (mr *MyRouter) HandleJoin(w http.ResponseWriter, r *http.Request) {
	respBodyBz, err := mr.handleJoin(r)
	if err != nil {
		log.Printf("failed to handle join: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(respBodyBz); err != nil {
		log.Printf("failed to write response body: %v", err)
		return
	}
}

func (mr *MyRouter) handleJoin(r *http.Request) ([]byte, error) {
	reqBodyBz, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	var reqBody JoinRequestBody
	if err := json.Unmarshal(reqBodyBz, &reqBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	reportBz, err := base64.StdEncoding.DecodeString(reqBody.ReportBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ReportBase64: %w", err)
	}

	report, err := enclave.VerifyRemoteReport(reportBz)
	// if err == attestation.ErrTCBLevelInvalid {
	// 	fmt.Printf("Warning: TCB level is invalid: %v\n%v\n", report.TCBStatus, tcbstatus.Explain(report.TCBStatus))
	// 	fmt.Println("We'll ignore this issue in this sample. For an app that should run in production, you must decide which of the different TCBStatus values are acceptable for you to continue.")
	// }
	if err != nil {
		return nil, fmt.Errorf("failed to verify report: %w", err)
	}

	log.Printf(
		"[REPORT RECEIVED] data:%v, securityVersion:%v, productID:%v, uniqueID:%v, singerID:%v",
		report.Data,
		report.SecurityVersion,
		binary.LittleEndian.Uint16(report.ProductID),
		report.UniqueID,
		hex.EncodeToString(report.SignerID),
	)

	if !bytes.Equal(report.Data[:len(sampleReportData)], sampleReportData) {
		log.Printf("%v", report.Data[:len(sampleReportData)])
		log.Printf("%v", sampleReportData)
		return nil, fmt.Errorf("invalid data in the report")
	}
	if report.SecurityVersion != 1 {
		return nil, fmt.Errorf("invalid security version in the report")
	}
	if binary.LittleEndian.Uint16(report.ProductID) != 1 {
		return nil, fmt.Errorf("invalid product ID in the report")
	}
	if hex.EncodeToString(report.SignerID) != "be9577a203acebd6957b180cc6ccd4a1a66d03e81657d7e7584ef469de5b9b99" {
		return nil, fmt.Errorf("invalid signer ID in the report")
	}
	//TODO: check unique ID

	peerNodePubKeyBz, err := base64.StdEncoding.DecodeString(reqBody.NodePublicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode NodePublicKeyBase64: %w", err)
	}
	peerNodePubKey, err := x509.ParsePKCS1PublicKey(peerNodePubKeyBz)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peerNodePubKey: %w", err)
	}
	encryptedOracleKey, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, peerNodePubKey, x509.MarshalPKCS1PrivateKey(mr.oraclePrivKey), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt oracle key: %w", err)
	}

	respBody := JoinResponseBody{
		EncryptedOracleKeyBase64: base64.StdEncoding.EncodeToString(encryptedOracleKey),
	}
	respBodyBz, err := json.Marshal(&respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body: %w", err)
	}
	return respBodyBz, nil
}
