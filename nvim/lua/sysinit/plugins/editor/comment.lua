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
          comment_line = '<C-_>',
          comment_visual = '<C-_>',
          textobject = '',
        },
      });
    end,
  },
}

return M

