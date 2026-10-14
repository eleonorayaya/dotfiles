package bar

import "fmt"

const (
	textFont      = "Rounded Mplus 1c"
	transparent   = "0x00000000"
	islandPadding = 12
	largeAppFont  = "sketchybar-app-font:Regular:19.0"
)

type islandSize struct {
	height   int
	radius   int
	pad      int
	fontSize int
}

var (
	smallIsland = islandSize{height: 28, radius: 14, pad: islandPadding, fontSize: 13}
	largeIsland = islandSize{height: 36, radius: 18, pad: 14, fontSize: 15}
	petRoomSize = islandSize{height: 44, radius: 22, pad: 14, fontSize: 15}
	tileSize    = islandSize{height: 30, radius: 15, pad: 12, fontSize: 15}
)

func font(style string, size int) string {
	return fmt.Sprintf("%s:%s:%d.0", textFont, style, size)
}

func barProps(place Placement) []string {
	return []string{
		P("height", place.Height),
		P("position", place.Position),
		"padding_left=0",
		"padding_right=0",
		P("y_offset", place.YOffset),
		P("color", transparent),
		"border_width=0",
		"blur_radius=0",
		"sticky=on",
		"font_smoothing=on",
	}
}

func defaultProps(c Colors) []string {
	return []string{
		P("icon.font", symbolFont+":Semibold:13.0"),
		P("icon.color", c.Text),
		"icon.padding_left=0",
		"icon.padding_right=6",
		P("label.font", font("Bold", smallIsland.fontSize)),
		P("label.color", c.Text),
		"label.padding_left=0",
		"label.padding_right=0",
		P("padding_left", itemPadding),
		P("padding_right", itemPadding),
		P("background.height", smallIsland.height),
		P("background.corner_radius", smallIsland.radius),
		"popup.background.border_width=2",
		P("popup.background.corner_radius", smallIsland.radius),
		P("popup.background.border_color", c.PopupBorder),
		P("popup.background.color", c.PopupBackground),
		"popup.blur_radius=20",
		"updates=when_shown",
	}
}

func bracketIsland(c Colors, sz islandSize) []string {
	return []string{
		"background.drawing=on",
		P("background.color", c.Island),
		P("background.border_color", c.IslandBorder),
		"background.border_width=2",
		P("background.height", sz.height),
		P("background.corner_radius", sz.radius),
	}
}

func islandOf(c Colors, sz islandSize) []string {
	return append(bracketIsland(c, sz),
		P("icon.color", c.Text),
		P("label.color", c.Text),
		P("label.font", font("Bold", sz.fontSize)),
		"background.padding_left=0",
		"background.padding_right=0",
		P("icon.padding_left", sz.pad),
		P("label.padding_right", sz.pad),
		P("padding_left", itemPadding),
		P("padding_right", itemPadding),
	)
}

func hotIslandOf(c Colors, sz islandSize) []string {
	return append(islandOf(c, sz),
		P("background.color", c.Hot),
		P("background.border_color", c.Hot),
		P("icon.color", c.HotText),
		P("label.color", c.HotText),
	)
}

func island(c Colors) []string {
	return islandOf(c, smallIsland)
}

func hotIsland(c Colors) []string {
	return hotIslandOf(c, smallIsland)
}

func labelOnly(sz islandSize) []string {
	return []string{"icon.drawing=off", P("label.padding_left", sz.pad)}
}

func faded(argb string) string {
	if len(argb) != 10 {
		return argb
	}
	return "0x55" + argb[4:]
}
