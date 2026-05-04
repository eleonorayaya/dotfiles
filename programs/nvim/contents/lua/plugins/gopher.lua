-- Go development enhancements
return {
  {
    "olexsmir/gopher.nvim",
    ft = { "go", "gomod", "gowork", "gotmpl" },
    dependencies = {
      "nvim-lua/plenary.nvim",
      "romus204/tree-sitter-manager.nvim",
    },
    config = function()
      require("gopher").setup({
        commands = {
          go = "go",
          gomodifytags = "gomodifytags",
          gotests = "gotests",
          impl = "impl",
          iferr = "iferr",
          dlv = "dlv",
        },
        goimports = "gopls",
        gofmt = "gopls",
      })
    end,
  },
}
