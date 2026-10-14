package bar

func Items(cfg Config, runner Runner, emitCmd string) []Item {
	items := []Item{
		NewWorkspaces(runner, cfg.WorkspaceIcons),
		NewFrontApp(runner),
		NewClock(),
	}
	if cfg.Bluetooth {
		items = append(items, NewBluetooth(runner, emitCmd))
	}
	items = append(items, NewNetwork(runner), NewSystem(runner))
	if cfg.MailApp != "" {
		items = append(items, NewMail(runner, cfg.MailApp))
	}
	if cfg.UtenaURL != "" {
		items = append(items, NewUtena(cfg.UtenaURL))
	}
	return items
}
