package bar

import "fmt"

const (
	textFont      = "Rounded Mplus 1c"
	islandHeight  = 28
	islandRadius  = 14
	islandPadding = 12

	largeHeight  = 36
	largeRadius  = 18
	largePadding = 14
	largeAppFont = "sketchybar-app-font:Regular:19.0"
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
		"color=0x00000000",
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
		P("label.font", font("Bold", 13)),
		P("label.color", c.Text),
		"label.padding_left=0",
		"label.padding_right=0",
		P("padding_left", itemPadding),
		P("padding_right", itemPadding),
		P("background.height", islandHeight),
		P("background.corner_radius", islandRadius),
		"popup.background.border_width=2",
		P("popup.background.corner_radius", islandRadius),
		P("popup.background.border_color", c.PopupBorder),
		P("popup.background.color", c.PopupBackground),
		"popup.blur_radius=20",
		"updates=when_shown",
	}
}

func island(c Colors) []string {
	return []string{
		"background.drawing=on",
		P("background.color", c.Island),
		P("background.border_color", c.IslandBorder),
		"background.border_width=2",
		P("background.height", islandHeight),
		P("background.corner_radius", islandRadius),
		P("icon.color", c.Text),
		P("label.color", c.Text),
		P("background.padding_left", 0),
		P("background.padding_right", 0),
		P("icon.padding_left", islandPadding),
		P("label.padding_right", islandPadding),
		P("padding_left", itemPadding),
		P("padding_right", itemPadding),
	}
}

func hotIsland(c Colors) []string {
	return append(island(c),
		P("background.color", c.Hot),
		P("background.border_color", c.Hot),
		P("icon.color", c.HotText),
		P("label.color", c.HotText),
	)
}

func faded(argb string) string {
	if len(argb) != 10 {
		return argb
	}
	return "0x55" + argb[4:]
}

func large() []string {
	return []string{
		P("background.height", largeHeight),
		P("background.corner_radius", largeRadius),
		P("label.font", font("Bold", 15)),
		P("icon.padding_left", largePadding),
		P("label.padding_left", largePadding),
		P("label.padding_right", largePadding),
	}
}

func bracketIsland(c Colors) []string {
	return []string{
		"background.drawing=on",
		P("background.color", c.Island),
		P("background.border_color", c.IslandBorder),
		"background.border_width=2",
		P("background.height", islandHeight),
		P("background.corner_radius", islandRadius),
	}
}
