package styles

type ThemeColors struct {
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
	AccentPeach           string
	AccentSalmon          string
	AccentPurple          string
	AccentLavender        string
	AccentGold            string
	AccentYellow          string
	AccentMint            string
	AccentBlue            string
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
