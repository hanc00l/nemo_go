package cert

import (
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"os"
	"path/filepath"
)

func GenerateSelfSignedCert(certFileName, keyFileName string) error {
	generator := SelfSignedCertGenerator{}

	artifacts, err := generator.Generate("127.0.0.1")
	if err != nil {
		return err
	}
	rootPath := conf.GetRootPath()
	if err = os.WriteFile(filepath.Join(rootPath, certFileName), artifacts.Cert, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(rootPath, keyFileName), artifacts.Key, 0644); err != nil {
		return err
	}

	return nil
}
