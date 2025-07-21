return {
	"rmagatti/auto-session",
	lazy = false,
	opts = {
		use_git_branch = true,
		auto_create = function()
			local cmd = "git rev-parse --is-inside-work-tree"
			return vim.fn.system(cmd) == "true\n"
		end,
		post_restore_cmds = {
			function()
				local nvim_tree_api = require("nvim-tree.api")

				nvim_tree_api.tree.change_root(vim.fn.getcwd())
				nvim_tree_api.tree.reload()
			end,
		},
	},
}
