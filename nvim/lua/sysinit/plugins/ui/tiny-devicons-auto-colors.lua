local M = {}

M.plugins = {
	{
		"rachartier/tiny-devicons-auto-colors.nvim",
		dependencies = {
			"nvim-tree/nvim-web-devicons",
			"rose-pine/neovim",
		},
		event = "VeryLazy",
		config = function()
			local theme_colors = require("rose-pine.palette")

			require("tiny-devicons-auto-colors").setup({
				colors = theme_colors,
			})
		end,
	},
}
return M

