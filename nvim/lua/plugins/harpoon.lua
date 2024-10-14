return {
	"ThePrimeagen/harpoon",
	branch = "harpoon2",
	dependencies = { "nvim-lua/plenary.nvim", "nvim-telescope/telescope.nvim" },
	config = function()
		local conf = require("telescope.config").values
		local telescope_state = require("telescope.actions.state")
		local telescope_finders = require("telescope.finders")

		local harpoon_get_paths = function(files)
			local paths = {}
			for _, item in ipairs(files.items) do
				table.insert(paths, item.value)
			end
			return paths
		end

		local function harpoon_make_finder(paths)
			return telescope_finders.new_table({ results = paths })
		end

		local harpoon = require("harpoon")
		harpoon:setup({})

		vim.keymap.set("n", "<leader>a", function()
			harpoon:list():add()
		end)

		local function toggle_telescope(harpoon_files)
			local file_paths = {}
			for _, item in ipairs(harpoon_files.items) do
				table.insert(file_paths, item.value)
			end

			require("telescope.pickers")
				.new({}, {
					prompt_title = "Harpoon",
					finder = harpoon_make_finder(file_paths),
					previewer = conf.file_previewer({}),
					sorter = conf.generic_sorter({}),
					attach_mappings = function(prompt_buffer_number, map)
						map("i", "<C-d>", function()
							local selected_entry = telescope_state.get_selected_entry()
							local current_picker = telescope_state.get_current_picker(prompt_buffer_number)

							table.remove(harpoon:list().items, selected_entry.index)
							current_picker:refresh(harpoon_make_finder(harpoon_get_paths(harpoon:list())))
						end)

						return true
					end,
				})
				:find()
		end

		vim.keymap.set("n", "<C-o>", function()
			toggle_telescope(harpoon:list())
		end, { desc = "Open harpoon window" })
	end,
}
