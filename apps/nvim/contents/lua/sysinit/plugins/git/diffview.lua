local M = {}

M.plugins = {
	{
		"sindrets/diffview.nvim",
		cmd = {
			"DiffviewOpen",
			"DiffviewClose",
			"DiffviewToggleFiles",
			"DiffviewFocusFiles",
			"DiffviewRefresh",
			"DiffviewFileHistory",
		},
		config = function()
			require("diffview").setup({})
		end,
		keys = {
			{
				"<localleader>gdo",
				"<CMD>DiffviewOpen<CR>",
				desc = "Open diffview",
			},
			{
				"<localleader>gdc",
				"<CMD>DiffviewClose<CR>",
				desc = "Close diffview",
			},
			{
				"<localleader>gdh",
				"<CMD>DiffviewFileHistory %<CR>",
				desc = "File history (current file)",
			},
			{
				"<localleader>gdH",
				"<CMD>DiffviewFileHistory<CR>",
				desc = "File history (all)",
			},
		},
	},
}

return M
