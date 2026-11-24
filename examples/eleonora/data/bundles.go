package data

import "github.com/eleonorayaya/shizuku/app"

func Bundles() map[string]app.Bundle {
	return map[string]app.Bundle{
		"go": {Plugins: []string{
			"gopls-lsp@claude-plugins-official",
			"bubbletea-tui@eleonorayaya-claude-code",
		}},
		"web":     {Plugins: []string{"typescript-lsp@claude-plugins-official"}},
		"rust":    {Plugins: []string{"rust-analyzer-lsp@claude-plugins-official"}},
		"lua":     {Plugins: []string{"lua-lsp@claude-plugins-official"}},
		"swift":   {Plugins: []string{"swift-lsp@claude-plugins-official"}},
		"homelab": {McpServers: []string{"homeassistant"}},
		"cad":     {McpServers: []string{"zoo", "bambuddy-local"}},
	}
}

func McpServers() map[string]map[string]any {
	return map[string]map[string]any{
		"homeassistant": {
			"type": "http",
			"url":  "https://homeassistant.home.tennyson.io/api/hass_mcp",
			"headers": map[string]any{
				"Authorization": "Bearer ${HOMEASSISTANT_TOKEN}",
			},
		},
		"zoo": {
			"type":    "stdio",
			"command": "uvx",
			"args":    []string{"zoo-mcp"},
		},
		"bambuddy-local": {
			"type":    "stdio",
			"command": "uvx",
			"args":    []string{"--with", "mcp==1.30.0", "bambuddy-mcp@0.2.0"},
			"env": map[string]any{
				"BAMBUDDY_URL":     "${BAMBUDDY_URL}",
				"BAMBUDDY_API_KEY": "${BAMBUDDY_API_KEY}",
			},
		},
	}
}
