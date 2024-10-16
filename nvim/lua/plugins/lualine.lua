return {
	"nvim-lualine/lualine.nvim",
	dependencies = { "nvim-tree/nvim-web-devicons" },
	config = function()
		local lualine = require("lualine")

		lualine.setup({
			options = {
				-- theme = theme,
				theme = "rose-pine",
			},
			sections = {
				lualine_b = {

					{
						"filetype",
						icon_only = true,
						separator = "",
						icon = { align = "right" },
						padding = {
							left = 1,
							right = 0,
						},
					},
					{
						"filename",
						path = 1,
						padding = {
							left = 0,
						},
					},
				},
				lualine_c = {},
				lualine_x = {},
				lualine_y = {},
				lualine_z = {},
			},
			inactive_sections = {
				lualine_b = {
					{
						"filename",
						path = 1,
					},
				},
				lualine_c = {},
				lualine_x = {},
			},
		})
	end,
}
