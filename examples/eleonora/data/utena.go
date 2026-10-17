package data

import "github.com/eleonorayaya/shizuku/programs/utena"

func Utena(browser string) utena.Options {
	return utena.Options{
		Browser:    browser,
		Bundles:    Bundles(),
		McpServers: McpServers(),
	}
}
