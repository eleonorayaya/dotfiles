return {
  {
    "nvim-mini/mini.comment",
    event = "VeryLazy",
    version = "*",
    config = function()
      require("mini.comment").setup({
        mappings = {
          comment = '',
          comment_line = '<C-/>',
          comment_visual = '<C-/>',
          textobject = '',
        },
      });
    end,
  },
}
