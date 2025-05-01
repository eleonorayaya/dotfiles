return {
	"williamboman/mason.nvim",
	dependencies = {
		"williamboman/mason-lspconfig.nvim",
		"WhoIsSethDaniel/mason-tool-installer.nvim",
	},
	config = function()
		-- import mason
		local mason = require("mason")

		-- import mason-lspconfig
		local mason_lspconfig = require("mason-lspconfig")

		local mason_tool_installer = require("mason-tool-installer")

		-- enable mason and configure icons
		mason.setup({
			ui = {
				icons = {
					package_installed = "✓",
					package_pending = "➜",
					package_uninstalled = "✗",
				},
			},
		})

		local enabled_lsps = {
			"lua_ls",
		}

		local enabled_tools = {
			"stylua",
		}

		if vim.g.btl_config.ruby_enabled then
			enabled_lsps = table.merge(enabled_lsps, {
				"ruby_lsp",
				"sorbet",
				"rubocop",
			})
		end

		if vim.g.btl_config.go_enabled then
			enabled_lsps = table.merge(enabled_lsps, {
				"gopls",
			})
		end

		if vim.g.btl_config.web_enabled then
			enabled_lsps = table.merge(enabled_lsps, {
				"ts_ls",
				"html",
				"cssls",
				"emmet_ls",
			})

			enabled_tools = table.merge(enabled_tools, {
				"prettier",
				"eslint_d",
			})
		end

		if vim.g.btl_config.python_enabled then
			enabled_lsps = table.merge(enabled_lsps, {
				"ruff",
			})

			enabled_tools = table.merge(enabled_tools, {
				-- "isort",
				-- "black",
				-- "pylint",
			})
		end

		mason_lspconfig.setup({
			ensure_installed = enabled_lsps,
		})

		mason_tool_installer.setup({
			ensure_installed = enabled_tools,
		})
	end,
}
