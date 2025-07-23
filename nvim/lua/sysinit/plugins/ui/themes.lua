local M = {}

M.plugins = {
	{
    {
      "rose-pine/neovim",
      priority = 1000,
      config = function()
        require("rose-pine").setup({
          variant = "moon", -- auto, main, moon, or dawn
          dark_variant = "moon", -- main, moon, or dawn
          dim_inactive_windows = false,
          extend_background_behind_borders = true,

          enable = {
            terminal = true,
            legacy_highlights = true,
            migrations = true,
          },

          styles = {
            bold = false;
            italic = true;
            transparency = true;
          },

          highlight_groups = {
            TelescopeBorder = { fg = "highlight_high", bg = "none" },
            TelescopeNormal = { bg = "none" },
            TelescopePromptNormal = { bg = "none" },
            TelescopeResultsNormal = { fg = "subtle", bg = "none" },
            TelescopeSelection = { fg = "text", bg = "none" },
            TelescopeSelectionCaret = { fg = "rose", bg = "rose" },
            Comment = { fg = "foam" },
            VertSplit = { fg = "muted", bg = "muted" },
          },
          before_highlight = function(group, highlight, palette)
            -- Change palette colour
            if highlight.fg == palette.pine then
              highlight.fg = palette.foam
            end
          end,
        })

        vim.cmd("colorscheme rose-pine-moon")

        vim.api.nvim_set_hl(0, "Normal", { bg = "none" })
        vim.api.nvim_set_hl(0, "NormalFloat", { bg = "none" })
      end,
    },
  },
}

return M

