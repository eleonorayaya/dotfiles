package styles

type Palette struct {
	Black         string
	Red           string
	Green         string
	Yellow        string
	Blue          string
	Magenta       string
	Cyan          string
	White         string
	BrightBlack   string
	BrightRed     string
	BrightGreen   string
	BrightYellow  string
	BrightBlue    string
	BrightMagenta string
	BrightCyan    string
	BrightWhite   string
}

type ThemeColors struct {
	Palette               Palette
	Surface               string
	SurfaceVariant        string
	SurfaceHighlight      string
	SurfaceBorder         string
	TextOnSurface         string
	TextOnSurfaceVariant  string
	TextOnSurfaceMuted    string
	TextOnSurfaceEmphasis string
	Primary               string
	PrimaryVariant        string
	TextOnPrimary         string
	Secondary             string
	TextOnSecondary       string
	Tertiary              string
	TertiaryVariant       string
	TextOnTertiary        string
	Accent                string
	Keyword               string
	String                string
	Number                string
	Constant              string
	Tag                   string
	Path                  string
	InlineCode            string
	Emphasis              string
	Added                 string
	Modified              string
	Removed               string
	Error                 string
	TextOnError           string
	Warning               string
	TextOnWarning         string
	Success               string
	TextOnSuccess         string
	Info                  string
	TextOnInfo            string
	Selection             string
	SelectionForeground   string
	Cursor                string
	CursorText            string
	Link                  string
	LinkHover             string
	Comment               string
}

type Theme struct {
	Name   string
	Type   string
	Colors ThemeColors
}

func (t Theme) WithDefaults() Theme {
	c := &t.Colors
	p := c.Palette
	defaults := []struct {
		field *string
		value string
	}{
		{&c.Accent, c.Primary},
		{&c.Keyword, p.Magenta},
		{&c.String, p.Green},
		{&c.Number, p.Yellow},
		{&c.Constant, p.Red},
		{&c.Tag, p.Blue},
		{&c.Path, p.Blue},
		{&c.InlineCode, p.Cyan},
		{&c.Emphasis, p.BrightMagenta},
		{&c.Added, p.Green},
		{&c.Modified, p.Yellow},
		{&c.Removed, p.Red},
	}
	for _, d := range defaults {
		if *d.field == "" {
			*d.field = d.value
		}
	}
	return t
}
