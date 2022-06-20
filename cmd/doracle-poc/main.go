package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/server"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	log "github.com/sirupsen/logrus"
)

func main() {
	pListenAddr := flag.String("laddr", "0.0.0.0:8080", "http listen addr")
	pInit := flag.Bool("init", false, "run doracle with the init mode")
	pPeer := flag.String("peer", "", "a peer addr for handshaking")
	flag.Parse()

	if *pInit && *pPeer != "" {
		log.Fatal("do not use -peer with -init")
	} else if *pInit {
		if err := mode.Init(); err != nil {
			log.Fatal("failed to run the init mode: %v", err)
		}
	} else if *pPeer != "" {
		if err := mode.Handshake(*pPeer); err != nil {
			log.Fatal("failed to run the handshake mode: %v", err)
		}
	}

	oraclePrivKeyBytes, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
	if err != nil {
		log.Fatalf("failed to load and unseal oracle key: %v", err)
	}

	srv := server.NewServer(secp256k1.PrivKeyFromBytes(oraclePrivKeyBytes))
	srvShutdownFunc := srv.ListenAndServe(*pListenAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	srvShutdownFunc()

	log.Info("terminating the process")
	os.Exit(0)
}
