package bar

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	BarTop    = "top"
	BarBottom = "bottom"

	BottomBinary    = "bottombar"
	topBinary       = "sketchybar"
	topBarHeight    = 40
	bottomBarHeight = 44

	MinShelfGap = bottomBarHeight

	centerDivisor       = 2
	bottomOffsetDivisor = centerDivisor * 2
)

type Spec struct {
	Socket    string
	Placement func(Config) Placement
	Items     func(Config, Runner) []Item
}

var specs = map[string]Spec{
	BarTop: {
		Socket:    "shizuku-bar.sock",
		Placement: topPlacement,
		Items:     Items,
	},
	BarBottom: {
		Socket:    "shizuku-bar-bottom.sock",
		Placement: bottomPlacement,
		Items:     BottomItems,
	},
}

func SpecFor(bar string) (Spec, error) {
	spec, ok := specs[bar]
	if !ok {
		return Spec{}, fmt.Errorf("unknown bar %q", bar)
	}
	return spec, nil
}

func (s Spec) SocketPath() string {
	return filepath.Join(os.TempDir(), s.Socket)
}

func topPlacement(cfg Config) Placement {
	return Placement{Bin: topBinary, Position: "top", Height: topBarHeight, YOffset: max((cfg.TopGap-topBarHeight)/centerDivisor, 0)}
}

func bottomPlacement(cfg Config) Placement {
	return Placement{Bin: BottomBinary, Position: "bottom", Height: bottomBarHeight, YOffset: max((cfg.BottomGap-bottomBarHeight)/bottomOffsetDivisor, 0)}
}
