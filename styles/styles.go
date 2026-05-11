package styles

type Gaps struct {
	InnerHorizontal int
	InnerVertical   int
	OuterLeft       int
	OuterBottom     int
	OuterTop        int
	OuterRight      int
}

type GapOverride struct {
	Pattern string
	Gaps    Gaps
}

type Styles struct {
	Theme       Theme
	WindowOpacity int
	Gaps          Gaps
	GapOverride   *GapOverride
}

type Option func(*Styles)

func WithTheme(t Theme) Option {
	return func(s *Styles) { s.Theme = t }
}

func WithWindowOpacity(opacity int) Option {
	return func(s *Styles) { s.WindowOpacity = opacity }
}

func WithGaps(g Gaps) Option {
	return func(s *Styles) { s.Gaps = g }
}

func WithGapOverride(pattern string, g Gaps) Option {
	return func(s *Styles) { s.GapOverride = &GapOverride{Pattern: pattern, Gaps: g} }
}

func New(opts ...Option) Styles {
	s := Styles{
		WindowOpacity: 85,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}
