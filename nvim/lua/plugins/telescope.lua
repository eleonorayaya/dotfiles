return {
	"nvim-telescope/telescope.nvim",
	dependencies = {
		"nvim-lua/plenary.nvim",
		{ "nvim-telescope/telescope-fzf-native.nvim", build = "make" },
		"nvim-tree/nvim-web-devicons",
	},
	config = function()
		local telescope = require("telescope")
		local actions = require("telescope.actions")

		telescope.setup({
			defaults = {
				path_display = { "smart" },
				mappings = {
					i = {
						["<C-k>"] = actions.move_selection_previous, -- move to prev result
						["<C-j>"] = actions.move_selection_next, -- move to next result
						["<C-q>"] = actions.send_selected_to_qflist + actions.open_qflist,
					},
				},
			},
		})

		telescope.load_extension("fzf")

		-- set keymaps
		local keymap = vim.keymap -- for conciseness

		keymap.set("n", "<leader>ff", "<cmd>Telescope find_files<cr>", { desc = "Fuzzy find files in cwd" })
		keymap.set("n", "<M-p>", "<cmd>Telescope git_files<cr>", { desc = "Fuzzy find git files" })
		keymap.set("n", "<leader>fr", "<cmd>Telescope oldfiles<cr>", { desc = "Fuzzy find recent files" })
		keymap.set("n", "<leader>fs", "<cmd>Telescope live_grep<cr>", { desc = "Find string in cwd" })
		keymap.set("n", "<leader>fc", "<cmd>Telescope grep_string<cr>", { desc = "Find string under cursor in cwd" })
	end,
}

-- return {
-- 	"nvim-telescope/telescope.nvim",
-- 	tag = "0.1.8",
-- 	dependencies = { "nvim-lua/plenary.nvim" },
-- 	config = function()
-- 		local builtin = require("telescope.builtin")
-- 		vim.keymap.set("n", "<leader>pf", builtin.find_files, { desc = "Telescope find files" })
-- 		vim.keymap.set("n", "<M-p>", builtin.git_files, { desc = "Telescope find files" })
-- 		vim.keymap.set("n", "<leader>ps", function()
-- 			builtin.grep_string({ search = vim.fn.input("Grep > ") })
-- 		end)
-- 	end,
-- }
