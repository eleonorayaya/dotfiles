package data

func PluginBundles() map[string][]string {
	return map[string][]string{
		"go": {
			"gopls-lsp@claude-plugins-official",
			"go@go-skills",
			"go-patterns@eleonorayaya-claude-code",
			"go-spec-reviewer@go-skills",
			"cobra-viper@go-skills",
			"fileflow-pathologize@go-skills",
			"go-release@go-skills",
			"bubbletea-tui@eleonorayaya-claude-code",
			"charm-dev@charm-dev-skills",
		},
		"web":        {"typescript-lsp@claude-plugins-official"},
		"rust":       {"rust-analyzer-lsp@claude-plugins-official"},
		"lua":        {"lua-lsp@claude-plugins-official"},
		"swift":      {"swift-lsp@claude-plugins-official"},
		"claude-dev": {"cc-toolkit@eleonorayaya-claude-code", "skill-creator@claude-plugins-official"},
		"docs":       {"notion-workspace-plugin@notion-plugin-marketplace"},
	}
}
