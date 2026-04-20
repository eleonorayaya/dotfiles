package app

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/util"
)

func VerifyInstallation(binaryName string) error {
	if !util.BinaryExists(binaryName) {
		return fmt.Errorf("%s not found in PATH after installation - please check installation logs", binaryName)
	}

	return nil
}
