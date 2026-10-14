package util

import (
	"fmt"
	"strconv"
	"strings"
)

func HexToARGB(hex string, alphaPercent int) string {
	hex = strings.TrimPrefix(hex, "#")
	alphaHex := fmt.Sprintf("%02X", (alphaPercent*255)/100)
	return "0x" + alphaHex + hex
}

func MixHex(a, b string, t float64) string {
	ca, okA := parseHexRGB(a)
	cb, okB := parseHexRGB(b)
	if !okA || !okB {
		return ""
	}
	var out [3]int
	for i := range out {
		out[i] = int(float64(ca[i]) + (float64(cb[i])-float64(ca[i]))*t + 0.5)
	}
	return fmt.Sprintf("#%02x%02x%02x", out[0], out[1], out[2])
}

func parseHexRGB(hex string) ([3]int, bool) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return [3]int{}, false
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return [3]int{}, false
	}
	return [3]int{int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff)}, true
}
