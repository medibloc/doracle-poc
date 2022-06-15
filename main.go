package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/gorilla/mux"
)

const keyFilePath = "/data/oracle-key.sealed"

func main() {
	pInit := flag.Bool("init", false, "run doracle with the init mode")
	pPeer := flag.String("peer", "", "a peer addr (do not use with -init)")
	flag.Parse()

	if *pInit && *pPeer != "" {
		log.Fatal("do not use -peer with -init")
	}

	if *pInit {
		if err := generateAndSaveOracleKey(); err != nil {
			log.Fatalf("failed to init: %v", err)
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
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening %s...", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func generateAndSaveOracleKey() error {
	oraclePrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate oracle key: %w", err)
	}

	oraclePrivKeyBz := x509.MarshalPKCS1PrivateKey(oraclePrivKey)
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

type MyRouter struct {
	oraclePrivKey *rsa.PrivateKey
}

func (mr *MyRouter) HandleJoin(w http.ResponseWriter, r *http.Request) {
	log.Printf("priv key: %v", mr.oraclePrivKey)
}
