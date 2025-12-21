local M = {}

M.plugins = {
  {
    "MeanderingProgrammer/render-markdown.nvim",
    dependencies = {
      "nvim-treesitter/nvim-treesitter",
      "nvim-tree/nvim-web-devicons",
    },
    ft = {
      "markdown",
      "octo",
    },
    cmd = {
    },
    config = function()
      require("render-markdown").setup({
        anti_conceal = {
          enabled = false,
        },
        code = {
          border = "thin",
          position = "left",
          language_icon = true,
        },
        latex = {
          enabled = false,
        },
        pipe_table = {
          above = "─",
          below = "─",
          border = {
            "╭",
            "┬",
            "╮",
            "├",
            "┼",
            "┤",
            "╰",
            "┴",
            "╯",
            "│",
            "─",
          },
        },
        completions = {
          lsp = {
            enabled = true,
          },
        },
        file_types = {
          "markdown",
          "octo",
        },
        sign = {
          enabled = false,
        },
      })
    end,
  },
}

return M

