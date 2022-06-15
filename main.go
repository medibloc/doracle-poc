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

func main() {
	pInit := flag.Bool("init", false, "run doracle with the init mode")
	pPeer := flag.String("peer", "", "a peer addr (do not use with -init)")
	flag.Parse()

	if *pInit && *pPeer != "" {
		log.Fatal("do not use -peer with -init")
	}

	if *pInit {
		if err := runInit(); err != nil {
			log.Fatalf("failed to init: %v", err)
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/join", handleJoin).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func runInit() error {
	oraclePrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate oracle key: %w", err)
	}

	oraclePrivKeyBz := x509.MarshalPKCS1PrivateKey(oraclePrivKey)
	sealed, err := ecrypto.SealWithProductKey(oraclePrivKeyBz, nil)
	if err != nil {
		return fmt.Errorf("failed to seal oracle key: %w", err)
	}

	if err := ioutil.WriteFile("oracle-key.sealed", sealed, 0644); err != nil {
		return fmt.Errorf("failed to write oracle-key.sealed: %w", err)
	}

	return nil
}

func handleJoin(w http.ResponseWriter, r *http.Request) {
	log.Println("join")
}
