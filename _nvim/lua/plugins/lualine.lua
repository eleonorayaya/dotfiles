return {
	"nvim-lualine/lualine.nvim",
	dependencies = { "nvim-tree/nvim-web-devicons" },
	config = function()
		local lualine = require("lualine")

		local custom_rosepine = require("lualine.themes.rose-pine")

		custom_rosepine.normal.b.bg = "none"
		custom_rosepine.normal.c.bg = "none"

		custom_rosepine.insert.b.bg = "none"
		custom_rosepine.insert.c.bg = "none"

		custom_rosepine.visual.b.bg = "none"
		custom_rosepine.visual.c.bg = "none"

		custom_rosepine.replace.b.bg = "none"
		custom_rosepine.replace.c.bg = "none"

		custom_rosepine.command.b.bg = "none"
		custom_rosepine.command.c.bg = "none"

		custom_rosepine.inactive.b.bg = "none"
		custom_rosepine.inactive.c.bg = "none"

		lualine.setup({
			options = {
				theme = custom_rosepine,
			},
			sections = {
				lualine_a = {},
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
