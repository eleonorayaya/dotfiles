local M = {}

M.plugins = {
	{
		"echasnovski/mini.comment",
		event = "BufReadPost",
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

return M

