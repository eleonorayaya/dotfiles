return {
	"nvim-telescope/telescope.nvim",
	branch = "0.1.x",
	dependencies = {
		"nvim-lua/plenary.nvim",
		{ "nvim-telescope/telescope-fzf-native.nvim", build = "make" },
		{
			"nvim-telescope/telescope-live-grep-args.nvim",
			-- This will not install any breaking changes.
			-- For major updates, this must be adjusted manually.
			version = "^1.0.0",
		},
		"nvim-tree/nvim-web-devicons",
	},
	config = function()
		local telescope = require("telescope")
		local actions = require("telescope.actions")
		local open_with_trouble = require("trouble.sources.telescope").open
		local lga_actions = require("telescope-live-grep-args.actions")
		local lga_shortcuts = require("telescope-live-grep-args.shortcuts")

		telescope.setup({
			defaults = {
				path_display = { "smart" },
				mappings = {
					i = {
						["<C-k>"] = actions.move_selection_previous, -- move to prev result
						["<C-j>"] = actions.move_selection_next, -- move to next result
						["<C-q>"] = actions.send_selected_to_qflist + actions.open_qflist,
						["<C-t>"] = open_with_trouble,
					},
				},
			},
			extensions = {
				live_grep_args = {
					auto_quoting = true,
					mappings = {
						i = {
							["<C-i>"] = lga_actions.quote_prompt({ postfix = " --iglob *" }),
							["<C-space>"] = actions.to_fuzzy_refine,
						},
					},
				},
			},
		})

		telescope.load_extension("fzf")
		telescope.load_extension("live_grep_args")

		-- set keymaps
		local keymap = vim.keymap -- for conciseness

		-- Find in all files
		keymap.set("n", "<leader>fa", "<cmd>Telescope find_files<cr>", { desc = "Fuzzy find files in cwd" })
		keymap.set("n", "<M-p>", "<cmd>Telescope git_files<cr>", { desc = "Fuzzy find git files" })
		keymap.set("n", "<leader>ff", function()
			telescope.extensions.live_grep_args.live_grep_args()
		end, { desc = "Find string in cwd" })

		-- Find visual in all files
		keymap.set("v", "<leader>fv", function()
			lga_shortcuts.grep_visual_selection({
				postfix = " --iglob *",
			})
		end, { desc = "Find visual selected string in cwd" })

		keymap.set("n", "<leader>bf", function()
			local opts = {}
			local curr_path = vim.fn.expand("%")
			opts["search_dirs"] = { curr_path }
			telescope.extensions.live_grep_args.live_grep_args(opts)
		end, { desc = "Find in current buffer" })

		keymap.set("v", "<leader>bv", function()
			lga_shortcuts.grep_word_visual_selection_current_buffer()
		end, { desc = "Find visual selected string in current buffer" })
	end,
}
