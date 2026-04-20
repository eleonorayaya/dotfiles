package styles

type Gaps struct {
	InnerHorizontal int
	InnerVertical   int
	OuterLeft       int
	OuterBottom     int
	OuterTop        int
	OuterRight      int
}

var (
	DesktopGaps = Gaps{
		InnerHorizontal: 16,
		InnerVertical:   16,
		OuterLeft:       64,
		OuterBottom:     128,
		OuterTop:        64,
		OuterRight:      64,
	}
	LaptopGaps = Gaps{
		InnerHorizontal: 8,
		InnerVertical:   8,
		OuterLeft:       8,
		OuterBottom:     8,
		OuterTop:        8,
		OuterRight:      8,
	}
)

type Styles struct {
	Theme         Theme
	WindowOpacity int
	Gaps          Gaps
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

func New(opts ...Option) Styles {
	s := Styles{
		WindowOpacity: 85,
		Gaps:          DesktopGaps,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}
