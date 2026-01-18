local M = {}

M.plugins = {
	{
		"folke/snacks.nvim",
		priority = 9800,
		lazy = false,
		dependencies = {
			"nvim-tree/nvim-web-devicons",
		},
		config = function()
			local snacks = require("snacks")
			snacks.setup({
				animate = {
					enabled = true,
					duration = 18,
					fps = 144,
				},
				bigfile = {
					enabled = true,
				},
				bufdelete = {
					enabled = true,
				},
				lazygit = {
					enabled = true,
					configure = false,
				},
				notifier = {
					enabled = true,
					margin = {
						top = 1,
						right = 1,
						bottom = 0,
					},
					style = "minimal",
					timeout = 1500,
				},
				picker = {
					enabled = true,
					ui_select = false,
					formatters = {
						d = {
							show_always = false,
							unselected = false,
						},
					},
					icons = {
						ui = {
							selected = " ",
							unselected = " ",
						},
					},
				},
				rename = {
					enabled = true,
				},
				scratch = {
					enabled = true,
				},
				statuscolumn = {
					enabled = false,
				},
				words = {
					enabled = true,
				},
				dashboard = {
					enabled = false,
				},
				debug = {
					enabled = false,
				},
				dim = {
					enabled = false,
				},
				explorer = {
					enabled = false,
				},
				git = {
					enabled = false,
				},
				gitbrowse = {
					enabled = false,
				},
				image = {
					enabled = false,
				},
				indent = {
					enabled = false,
				},
				layout = {
					enabled = false,
				},
				profiler = {
					enabled = true,
				},
				quickfile = {
					enabled = true,
				},
				terminal = {
					enabled = true,
				},
				scope = {
					enabled = false,
				},
				scroll = {
					enabled = false,
				},
				toggle = {
					enabled = false,
				},
				win = {
					enabled = false,
				},
				zen = {
					enabled = false,
				},
			})

			vim.notify = snacks.notifier
			vim.ui.input = snacks.input
		end,
		keys = function()
			local snacks = require("snacks")
			local default_keys = {
				{
					"<leader>bs",
					function()
						snacks.scratch()
					end,
					desc = "Toggle scratchpad",
				},
				{
					"<localleader>gg",
					function()
						snacks.lazygit()
					end,
					desc = "Open LazyGit UI",
				},
				{
					"<leader>ns",
					function()
						snacks.notifier.show_history()
					end,
					desc = "Show",
				},
				{
					"<leader>nc",
					function()
						snacks.notifier.hide()
					end,
					desc = "Dismiss",
				},
			}

			if vim.env.SYSINIT_DEBUG ~= "1" then
				return default_keys
			end

			local debug_keys = {
				{
					"<localleader>px",
					function()
						snacks.profiler.stop()
					end,
					desc = "Stop Profiler",
				},
				{
					"<localleader>pf",
					function()
						snacks.profiler.pick()
					end,
					desc = "Profiler Picker",
				},
				{
					"<localleader>pp",
					function()
						snacks.toggle.profiler()
					end,
					desc = "Toggle Profiler",
				},
				{
					"<localleader>ph",
					function()
						snacks.toggle.profiler_highlights()
					end,
					desc = "Toggle Profiler Highlights",
				},
				{
					"<localleader>ps",
					function()
						snacks.profiler.scratch()
					end,
					desc = "Profiler Scratch Buffer",
				},
			}

			for _, key in ipairs(debug_keys) do
				table.insert(default_keys, key)
			end

			return default_keys
		end,
	},
}

return M
