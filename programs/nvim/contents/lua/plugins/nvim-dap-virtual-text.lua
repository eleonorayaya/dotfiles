return {
  {
    "theHamsta/nvim-dap-virtual-text",
    dependencies = {
      "romus204/tree-sitter-manager.nvim",
    },
    lazy = true,
    config = function()
      require("nvim-dap-virtual-text").setup()
    end,
  },
}
