return {
	"jghauser/fold-cycle.nvim",
	config = function()
		require("fold-cycle").setup()

		vim.keymap.set("n", "<leader>zo", function()
			require("fold-cycle").open_all()
		end, { desc = "Cycle through folds" })

		vim.keymap.set("n", "<leader>zc", function()
			require("fold-cycle").close_all()
		end, { desc = "Close all folds" })

		vim.keymap.set("n", "<leader>zz", function()
			require("fold-cycle").toggle_all()
		end, { desc = "Toggle all folds" })
	end,
}
