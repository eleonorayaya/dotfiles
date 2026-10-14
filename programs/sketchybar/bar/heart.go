package bar

import "context"

const heartItem = "heart"

type Heart struct {
	colors Colors
}

func NewHeart(colors Colors) *Heart {
	return &Heart{colors: colors}
}

func (h *Heart) Setup(ctx context.Context, b *Batch) error {
	b.Add(heartItem, "left", append(island(h.colors),
		P("icon", iconHeart),
		P("icon.font", font("ExtraBold", 15)),
		P("icon.color", h.colors.Hot),
		P("icon.padding_right", islandPadding),
		"label.drawing=off",
	)...)
	return nil
}
