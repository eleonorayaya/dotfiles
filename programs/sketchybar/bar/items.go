package bar

func Items(cfg Config, runner Runner) []Item {
	items := []Item{
		NewWorkspaces(runner),
		NewFrontApp(runner),
		NewClock(),
		NewBluetooth(runner),
		NewNetwork(runner),
		NewSystem(runner),
	}
	if cfg.MailApp != "" {
		items = append(items, NewMail(runner, cfg.MailApp))
	}
	return append(items, NewUtena(utenaURL))
}
