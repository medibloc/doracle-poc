package node

import (
	"fmt"
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func CmdReadEncryptedFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read-encrypt-data [file-path]",
		Short: "read encrypted data from file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			if filePath == "" {
				return fmt.Errorf("file-path is empty")
			}

			encryptedData, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			oraclePrivateKeyBz, err := sgx.UnsealFromFile(mode.OracleKeyFilePath)
			if err != nil {
				return err
			}

			oraclePrivateKey := secp256k1.PrivKeyFromBytes(oraclePrivateKeyBz)
			data, err := secp256k1.Decrypt(oraclePrivateKey, encryptedData)
			if err != nil {
				return err
			}

			log.Println("decrypted data: ", string(data))

			return nil
		},
	}

	return cmd
}
