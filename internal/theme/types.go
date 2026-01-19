package theme

import "fmt"

type ThemeColors struct {
	Surface               string `json:"surface"`
	SurfaceVariant        string `json:"surfaceVariant"`
	SurfaceHighlight      string `json:"surfaceHighlight"`
	SurfaceBorder         string `json:"surfaceBorder"`
	TextOnSurface         string `json:"textOnSurface"`
	TextOnSurfaceVariant  string `json:"textOnSurfaceVariant"`
	TextOnSurfaceMuted    string `json:"textOnSurfaceMuted"`
	TextOnSurfaceEmphasis string `json:"textOnSurfaceEmphasis"`
	Primary               string `json:"primary"`
	PrimaryVariant        string `json:"primaryVariant"`
	TextOnPrimary         string `json:"textOnPrimary"`
	Secondary             string `json:"secondary"`
	TextOnSecondary       string `json:"textOnSecondary"`
	Tertiary              string `json:"tertiary"`
	TertiaryVariant       string `json:"tertiaryVariant"`
	TextOnTertiary        string `json:"textOnTertiary"`
	Accent                string `json:"accent"`
	AccentPeach           string `json:"accentPeach"`
	AccentSalmon          string `json:"accentSalmon"`
	AccentPurple          string `json:"accentPurple"`
	AccentLavender        string `json:"accentLavender"`
	AccentGold            string `json:"accentGold"`
	AccentYellow          string `json:"accentYellow"`
	AccentMint            string `json:"accentMint"`
	AccentBlue            string `json:"accentBlue"`
	Error                 string `json:"error"`
	TextOnError           string `json:"textOnError"`
	Warning               string `json:"warning"`
	TextOnWarning         string `json:"textOnWarning"`
	Success               string `json:"success"`
	TextOnSuccess         string `json:"textOnSuccess"`
	Info                  string `json:"info"`
	TextOnInfo            string `json:"textOnInfo"`
	Selection             string `json:"selection"`
	SelectionForeground   string `json:"selectionForeground"`
	Cursor                string `json:"cursor"`
	CursorText            string `json:"cursorText"`
	Link                  string `json:"link"`
	LinkHover             string `json:"linkHover"`
	Comment               string `json:"comment"`
}

type Theme struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Colors ThemeColors `json:"colors"`
}

func (c *ThemeColors) Validate() error {
	if c.Surface == "" {
		return fmt.Errorf("missing required color: surface")
	}
	if c.SurfaceVariant == "" {
		return fmt.Errorf("missing required color: surfaceVariant")
	}
	if c.SurfaceHighlight == "" {
		return fmt.Errorf("missing required color: surfaceHighlight")
	}
	if c.SurfaceBorder == "" {
		return fmt.Errorf("missing required color: surfaceBorder")
	}
	if c.TextOnSurface == "" {
		return fmt.Errorf("missing required color: textOnSurface")
	}
	if c.TextOnSurfaceVariant == "" {
		return fmt.Errorf("missing required color: textOnSurfaceVariant")
	}
	if c.TextOnSurfaceMuted == "" {
		return fmt.Errorf("missing required color: textOnSurfaceMuted")
	}
	if c.TextOnSurfaceEmphasis == "" {
		return fmt.Errorf("missing required color: textOnSurfaceEmphasis")
	}
	if c.Primary == "" {
		return fmt.Errorf("missing required color: primary")
	}
	if c.PrimaryVariant == "" {
		return fmt.Errorf("missing required color: primaryVariant")
	}
	if c.TextOnPrimary == "" {
		return fmt.Errorf("missing required color: textOnPrimary")
	}
	if c.Secondary == "" {
		return fmt.Errorf("missing required color: secondary")
	}
	if c.TextOnSecondary == "" {
		return fmt.Errorf("missing required color: textOnSecondary")
	}
	if c.Tertiary == "" {
		return fmt.Errorf("missing required color: tertiary")
	}
	if c.TertiaryVariant == "" {
		return fmt.Errorf("missing required color: tertiaryVariant")
	}
	if c.TextOnTertiary == "" {
		return fmt.Errorf("missing required color: textOnTertiary")
	}
	if c.Accent == "" {
		return fmt.Errorf("missing required color: accent")
	}
	if c.AccentPeach == "" {
		return fmt.Errorf("missing required color: accentPeach")
	}
	if c.AccentSalmon == "" {
		return fmt.Errorf("missing required color: accentSalmon")
	}
	if c.AccentPurple == "" {
		return fmt.Errorf("missing required color: accentPurple")
	}
	if c.AccentLavender == "" {
		return fmt.Errorf("missing required color: accentLavender")
	}
	if c.AccentGold == "" {
		return fmt.Errorf("missing required color: accentGold")
	}
	if c.AccentYellow == "" {
		return fmt.Errorf("missing required color: accentYellow")
	}
	if c.AccentMint == "" {
		return fmt.Errorf("missing required color: accentMint")
	}
	if c.AccentBlue == "" {
		return fmt.Errorf("missing required color: accentBlue")
	}
	if c.Error == "" {
		return fmt.Errorf("missing required color: error")
	}
	if c.TextOnError == "" {
		return fmt.Errorf("missing required color: textOnError")
	}
	if c.Warning == "" {
		return fmt.Errorf("missing required color: warning")
	}
	if c.TextOnWarning == "" {
		return fmt.Errorf("missing required color: textOnWarning")
	}
	if c.Success == "" {
		return fmt.Errorf("missing required color: success")
	}
	if c.TextOnSuccess == "" {
		return fmt.Errorf("missing required color: textOnSuccess")
	}
	if c.Info == "" {
		return fmt.Errorf("missing required color: info")
	}
	if c.TextOnInfo == "" {
		return fmt.Errorf("missing required color: textOnInfo")
	}
	if c.Selection == "" {
		return fmt.Errorf("missing required color: selection")
	}
	if c.SelectionForeground == "" {
		return fmt.Errorf("missing required color: selectionForeground")
	}
	if c.Cursor == "" {
		return fmt.Errorf("missing required color: cursor")
	}
	if c.CursorText == "" {
		return fmt.Errorf("missing required color: cursorText")
	}
	if c.Link == "" {
		return fmt.Errorf("missing required color: link")
	}
	if c.LinkHover == "" {
		return fmt.Errorf("missing required color: linkHover")
	}
	if c.Comment == "" {
		return fmt.Errorf("missing required color: comment")
	}
	return nil
}
