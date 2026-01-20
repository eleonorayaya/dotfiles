local M = {}

function _G.harpoon_add_file()
	local n = vim.api.nvim_buf_get_name(vim.api.nvim_get_current_buf())

	if string.match(n, "NvimTree") then
		return
	end

  local harpoon = require("harpoon")
  harpoon:list():add()
end

function _G.harpoon_open_telescope_marks()
  local conf = require("telescope.config").values
  local harpoon = require("harpoon")

  local harpoon_files = harpoon:list()
  local file_paths = {}
  for _, item in ipairs(harpoon_files.items) do
    table.insert(file_paths, item.value)
  end

  require("telescope.pickers").new({}, {
    prompt_title = "Harpoon",
    finder = require("telescope.finders").new_table({
      results = file_paths,
    }),
    previewer = conf.file_previewer({}),
    sorter = conf.generic_sorter({}),
  }):find()
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
        },
        -- default = {
        --   display = function(item)
        --     put(item)
        --     return item.value
        --   end,
        -- }
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

