local M = {}

M.plugins = {
  {
    {
      "AmberLehmann/candyland.nvim",
      priority = 1000,
      config = function()
        vim.cmd("colorscheme candyland")

        vim.api.nvim_set_hl(0, "Normal", { bg = "none" })
        vim.api.nvim_set_hl(0, "NormalFloat", { bg = "none" })
      end,
    },
  },
}

return M

