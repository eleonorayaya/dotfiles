package bar

func Items(cfg Config, runner Runner) []Item {
	c := cfg.Colors
	mood := NewMood()
	items := []Item{
		NewLayout(runner, cfg.Margin, cfg.MarginOverride, ""),
		NewHeart(c),
		NewWorkspace(runner, c),
		NewFrontApp(runner, c),
		NewKaomoji(mood, c),
		NewClock(c),
		NewSystem(runner, c, mood),
		NewBluetooth(runner, c),
		NewNetwork(runner, c, mood),
	}
	if len(cfg.Badges) > 0 {
		items = append(items, NewBadges(runner, c, cfg.Badges, mood))
	}
	return append(items, NewUtena(utenaURL, c))
}
