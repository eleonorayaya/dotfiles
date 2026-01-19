package shizukuconfig

type StylesConfig struct {
	ThemeName string `yaml:"theme"`
}

type Styles struct {
	Theme *Theme
}
