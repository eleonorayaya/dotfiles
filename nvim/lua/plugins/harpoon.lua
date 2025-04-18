return {
	"ThePrimeagen/harpoon",
	branch = "harpoon2",
	opts = {
		settings = {},
	},
	dependencies = { "nvim-lua/plenary.nvim", "nvim-telescope/telescope.nvim" },
	config = function()
		local conf = require("telescope.config").values
		local telescope_state = require("telescope.actions.state")
		local telescope_finders = require("telescope.finders")

		local is_git_dir = function()
			local out = vim.fn.systemlist("git rev-parse --is-inside-work-tree")

			if vim.v.shell_error ~= 0 then
				print(string.format("git failed with: %s", table.concat(out, "\n")))
				return false
			end

			return out[1] == "true"
		end

		local get_current_branch = function()
			if not is_git_dir() then
				return ""
			end

			local out = vim.fn.systemlist("git rev-parse --abbrev-ref HEAD")

			if vim.v.shell_error ~= 0 then
				print(string.format("git failed with: %s", table.concat(out, "\n")))
				return ""
			end

			return out[1]
		end

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
		harpoon.setup({
			settings = {
				save_on_toggle = true,
				sync_on_ui_close = true,
			},
		})

		vim.keymap.set("n", "<leader>a", function()
			local n = vim.api.nvim_buf_get_name(vim.api.nvim_get_current_buf())
			if string.match(n, "NvimTree") then
				return
			end

			local c = get_current_branch()
			require("harpoon"):list(c):add()
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

							local c = get_current_branch()
							table.remove(harpoon:list(c).items, selected_entry.index)
							current_picker:refresh(harpoon_make_finder(harpoon_get_paths(harpoon:list(c))))
						end)

						return true
					end,
				})
				:find()
		end

		vim.keymap.set("n", "<C-o>", function()
			local c = get_current_branch()
			toggle_telescope(harpoon:list(c))
		end, { desc = "Open harpoon window" })
	end,
}
