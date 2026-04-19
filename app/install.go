package app

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

type Installer interface {
	Install(cfg *config.Config) error
}

func VerifyInstallation(binaryName string) error {
	if !util.BinaryExists(binaryName) {
		return fmt.Errorf("%s not found in PATH after installation - please check installation logs", binaryName)
	}

	return nil
}
