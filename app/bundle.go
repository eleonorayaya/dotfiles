package app

type Bundle struct {
	Plugins    []string `json:"plugins,omitempty"`
	McpServers []string `json:"mcp_servers,omitempty"`
}
