package sgx

import (
	"fmt"
	"github.com/edgelesssys/ego/ecrypto"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

/*//If you want is not SGX environment execute, you can use below methods.

func SealToFile(data []byte, filePath string) error {
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is written successfully", filePath)

	return nil
}

func UnsealFromFile(filePath string) ([]byte, error) {
	sealed, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return sealed, nil
}*/

func SealToFile(data []byte, filePath string) error {
	sealed, err := ecrypto.SealWithProductKey(data, nil)
	if err != nil {
		return fmt.Errorf("failed to seal oracle key: %w", err)
	}

	if err := ioutil.WriteFile(filePath, sealed, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	log.Infof("%s is written successfully", filePath)

	return nil
}

func UnsealFromFile(filePath string) ([]byte, error) {
	sealed, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	key, err := ecrypto.Unseal(sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal oracle key: %w", err)
	}

	return key, nil
}
