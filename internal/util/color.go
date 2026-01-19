package util

import (
	"fmt"
	"strings"
)

func HexToARGB(hex string, alphaPercent int) string {
	hex = strings.TrimPrefix(hex, "#")
	alphaHex := fmt.Sprintf("%02X", (alphaPercent*255)/100)
	return "0x" + alphaHex + hex
}
