package shizukuapp

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type Installer interface {
	Install(config *shizukuconfig.Config) error
}

func VerifyInstallation(binaryName string) error {
	if !util.BinaryExists(binaryName) {
		return fmt.Errorf("%s not found in PATH after installation - please check installation logs", binaryName)
	}

	return nil
}

