package bar

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	BarTop    = "top"
	BarBottom = "bottom"

	topBinary       = "sketchybar"
	BottomBinary    = "bottombar"
	topBarHeight    = 40
	bottomBarHeight = 44

	bottomOffsetDivisor = 4
)

func TopPlacement() Placement {
	return Placement{Bin: topBinary, Position: "top", Height: topBarHeight, YOffset: 6}
}

func BottomPlacement(cfg Config) Placement {
	return Placement{Bin: BottomBinary, Position: "bottom", Height: bottomBarHeight, YOffset: max((cfg.BottomGap-bottomBarHeight)/bottomOffsetDivisor, 0)}
}

func PlacementFor(bar string, cfg Config) (Placement, error) {
	switch bar {
	case BarTop:
		return TopPlacement(), nil
	case BarBottom:
		return BottomPlacement(cfg), nil
	}
	return Placement{}, fmt.Errorf("unknown bar %q", bar)
}

func DefaultSocketPath(bar string) string {
	name := "shizuku-bar.sock"
	if bar == BarBottom {
		name = "shizuku-bar-bottom.sock"
	}
	return filepath.Join(os.TempDir(), name)
}

func ItemsFor(bar string, cfg Config, runner Runner) []Item {
	if bar == BarBottom {
		return BottomItems(cfg, runner)
	}
	return Items(cfg, runner)
}
