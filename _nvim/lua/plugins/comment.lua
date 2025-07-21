return {
	"numToStr/Comment.nvim",
	event = { "BufReadPre", "BufNewFile" },
	dependencies = {
		"JoosepAlviste/nvim-ts-context-commentstring",
	},
	config = function()
		-- import comment plugin safely
		local comment = require("Comment")
		local api = require("Comment.api")
		local ts_context_commentstring = require("ts_context_commentstring.integrations.comment_nvim")

		-- enable comment
		comment.setup({
			-- for commenting tsx, jsx, svelte, html files
			pre_hook = ts_context_commentstring.create_pre_hook(),
		})

		vim.keymap.set("n", "<C-_>", api.toggle.linewise.current)

		local esc = vim.api.nvim_replace_termcodes("<ESC>", true, false, true)

		vim.keymap.set("x", "<C-_>", function()
			vim.api.nvim_feedkeys(esc, "nx", false)
			api.toggle.linewise(vim.fn.visualmode())
		end)
	end,
}
