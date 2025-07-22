local M = {}

function _G.harpoon_add_file()
	local n = vim.api.nvim_buf_get_name(vim.api.nvim_get_current_buf())

	if string.match(n, "NvimTree") then
		return
	end

  local harpoon = require("harpoon")
  harpoon:list():add()
end

function _G.harpoon_get_paths(files)
	local paths = {}

	for _, item in ipairs(files.items) do
		table.insert(paths, item.value)
	end

	return paths
end

function _G.harpoon_make_finder(paths)
	return require("telescope.finders").new_table({ results = paths })
end

function _G.harpoon_open_telescope_marks()
	require("telescope").extensions.harpoon.marks()
end

function _G.harpoon_clear_marks()
  local harpoon = require("harpoon")
  harpoon:list():clear()
end

M.plugins = {
  {
    "ThePrimeagen/harpoon",
    branch = "harpoon2",
    dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-telescope/telescope.nvim"
    },
    config = function()
      local harpoon = require("harpoon")
      harpoon.setup({
        config = {
          save_on_toggle = true,
          sync_on_ui_close = true,

        }
      })

      local harpoon_extensions = require("harpoon.extensions")
      harpoon:extend(harpoon_extensions.builtins.highlight_current_file())
    end,
		keys = {
			{
				"<leader>a",
				function()
          harpoon_add_file()
				end,
				desc = "Harpoon: Add file",
			},
      {
				"<C-o>",
				function()
          harpoon_open_telescope_marks()
				end,
				desc = "Harpoon: Add file",
			},
      {
        "<localleader>hc",
        function()
          harpoon_clear_marks()
        end,
        desc = "Harpoon: Clear"
      }
    },
  },
}

return M

