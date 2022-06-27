package cmd

import (
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/server"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func CmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			pListenAddr, err := cmd.Flags().GetString("laddr")
			if err != nil {
				return err
			}
			pInit, err := cmd.Flags().GetBool("init")
			if err != nil {
				return err
			}
			pPeer, err := cmd.Flags().GetString("peer")
			if err != nil {
				return err
			}

			if pInit && pPeer != "" {
				log.Fatal("do not use -peer with -init")
			} else if pInit {
				if err := mode.Init(); err != nil {
					log.Fatal("failed to run the init mode: %w", err)
				}
			} else if pPeer != "" {
				if err := mode.Handshake(pPeer); err != nil {
					log.Fatal("failed to run the handshake mode: %w", err)
				}
			}

			oraclePrivKeyBytes, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
			if err != nil {
				log.Fatalf("failed to load and unseal oracle key: %v", err)
			}

			srv := server.NewServer(secp256k1.PrivKeyFromBytes(oraclePrivKeyBytes))
			srvShutdownFunc := srv.ListenAndServe(pListenAddr)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
			<-sigCh

			srvShutdownFunc()

			log.Info("terminating the process")
			os.Exit(0)

			return nil
		},
	}

	cmd.Flags().String("laddr", "0.0.0.0:8080", "http listen addr")
	cmd.Flags().Bool("init", false, "run doracle with the init mode")
	cmd.Flags().String("peer", "", "a peer addr for handshaking")
	return cmd
}
