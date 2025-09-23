local nvim_config = require("sysinit.config.nvim_config").load_config()
local M = {}

M.plugins = {
{
  "mfussenegger/nvim-dap",
  dependencies = {
    "suketa/nvim-dap-ruby"
  },
  config = function()
    require("dap-ruby").setup()
  end
}
}

return M

