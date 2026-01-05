local M = {}

M.plugins = {
  {
    {
      dir = vim.fn.stdpath("config") .. "/colors",
      priority = 1000,
      config = function()
        vim.cmd("colorscheme monade")
      end,
    },
  },
}

return M

