package mode

import (
	"github.com/medibloc/doracle-poc/pkg/secp256k1"
	"github.com/medibloc/doracle-poc/pkg/sgx"
	log "github.com/sirupsen/logrus"
)

const OracleKeyFilePath = "/data/oracle-key.sealed"

func Init() error {
	oraclePrivKey, err := secp256k1.NewPrivKey()
	if err != nil {
		log.Fatalf("failed to generate oracle key: %v", err)
	}

	if err := sgx.SealToFile(oraclePrivKey.Serialize(), OracleKeyFilePath); err != nil {
		log.Fatalf("failed to save oracle key: %v", err)
	}

	return nil
}
