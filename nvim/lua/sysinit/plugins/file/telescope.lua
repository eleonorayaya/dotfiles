local M = {}

function _G.telescope_pick_project_files()
	local opts = {}

	if is_git_repo() then
		opts = {
			cwd = get_git_root(),
		}
	end

	require("telescope.builtin").find_files(opts)
end

M.plugins = {
	{
		"nvim-telescope/telescope.nvim",
		branch = "master",
		lazy = false,
		dependencies = {
			"debugloop/telescope-undo.nvim",
			"nvim-lua/plenary.nvim",
			"nvim-telescope/telescope-dap.nvim",
			"nvim-telescope/telescope-fzy-native.nvim",
			"nvim-telescope/telescope-live-grep-args.nvim",
			"nvim-telescope/telescope-ui-select.nvim",
			"nvim-tree/nvim-web-devicons",
			"olimorris/persisted.nvim",
			"nvim-treesitter/nvim-treesitter",
		},
		config = function()
			local telescope = require("telescope")
			local actions = require("telescope.actions")
			local themes = require("telescope.themes")
      local lga_actions = require("telescope-live-grep-args.actions")
      local lga_shortcuts = require("telescope-live-grep-args.shortcuts")

			telescope.setup({
				defaults = {
					prompt_prefix = "   ",
          path_display = {
            truncate = {
              len = 2,
              exclude = {
                1,
                2,
                -1,
                -2,
              }
            }
          },
					selection_caret = "",
					entry_prefix = "",
					-- Enhanced border styling
					borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" },
					results_title = "",
					prompt_title = "",
					preview_title = "",
					sorting_strategy = "ascending",
					layout_config = {
						horizontal = {
							prompt_position = "top",
							-- preview_width = 0.55,
						},
						width = 0.87,
						height = 0.80,
					},
					mappings = {
						n = {
							["q"] = actions.close,
							["<Tab>"] = actions.move_selection_next,
							["<S-Tab>"] = actions.move_selection_previous,
							["j"] = actions.move_selection_next,
							["k"] = actions.move_selection_previous,
							["<Down>"] = actions.move_selection_next,
							["<Up>"] = actions.move_selection_previous,
							["<CR>"] = actions.select_default,
							["<localleader>s"] = actions.select_horizontal,
							["<localleader>v"] = actions.select_vertical,
							["<localleader>t"] = actions.select_tab,
						},
						i = {
							["<Tab>"] = actions.move_selection_next,
							["<S-Tab>"] = actions.move_selection_previous,
							["<C-j>"] = actions.move_selection_next,
							["<C-k>"] = actions.move_selection_previous,
							["<Down>"] = actions.move_selection_next,
							["<Up>"] = actions.move_selection_previous,
							["<CR>"] = actions.select_default,
							["<C-u>"] = actions.preview_scrolling_up,
							["<C-d>"] = actions.preview_scrolling_down,
						},
					},
					file_ignore_patterns = {
						"%.git",
						"%.cache",
						"%.png",
						"%.jpg",
						"%.jpeg",
						"%.o",
						".cache",
						"Build",
            "sorbet",
            "node_modules"
					},
				},
				extensions = {
					["ui-select"] = {
						themes.get_dropdown(),
					},
					persisted = {
						layout_config = { width = 0.55, height = 0.55 },
					},
					fzy_native = {
						override_generic_sorter = true,
						override_file_sorter = true,
					},
					dap = {},
					live_grep_args = {},
					undo = {
						side_by_side = true,
						layout_strategy = "vertical",
						layout_config = {
							preview_height = 0.8,
						},
					},
				},
				pickers = {
					find_files = {
						hidden = true,
					},
					live_grep = {
						additional_args = function()
							return { "--hidden" }
						end,
					},
					colorscheme = {
						enable_preview = true,
					},
				},
				vimgrep_arguments = {
					"rg",
					"--color=never",
					"--no-heading",
					"--hidden",
					"--with-filename",
					"--line-number",
					"--column",
					"--smart-case",
					"--trim",
				},
			})

			-- Extensions will be loaded on demand via pickers
			local function lazy_load_ext(ext)
				local ok, _ = pcall(telescope.load_extension, ext)
				if not ok then
					return
				end
			end
			vim.api.nvim_create_autocmd("User", {
				pattern = "TelescopeFindFiles",
				callback = function()
					lazy_load_ext("fzy_native")
				end,
			})
			vim.api.nvim_create_autocmd("User", {
				pattern = "TelescopeLiveGrep",
				callback = function()
					lazy_load_ext("live_grep_args")
				end,
			})
			vim.api.nvim_create_autocmd("User", {
				pattern = "TelescopeUndo",
				callback = function()
					lazy_load_ext("undo")
				end,
			})
			vim.api.nvim_create_autocmd("User", {
				pattern = "TelescopeDap",
				callback = function()
					lazy_load_ext("dap")
				end,
			})
			vim.api.nvim_create_autocmd("User", {
				pattern = "TelescopeUiSelect",
				callback = function()
					lazy_load_ext("ui-select")
				end,
			})
			vim.api.nvim_create_autocmd("User", {
				pattern = "TelescopePersisted",
				callback = function()
					lazy_load_ext("persisted")
				end,
			})
		end,
		keys = function()
			local actions = require("telescope.actions")
      local lga_actions = require("telescope-live-grep-args.actions")
      local lga_shortcuts = require("telescope-live-grep-args.shortcuts")

			return {
				{
					"<M-p>",
					function()
						telescope_pick_project_files()
					end,
					desc = "Git Files",
				},
        {
          "<leader>ff",
          function()
            require("telescope").extensions.live_grep_args.live_grep_args({
            auto_quoting = true,
            mappings = {
              i = {
                ["<C-i>"] = lga_actions.quote_prompt({ postfix = " --iglob *" }),
                ["<C-space>"] = actions.to_fuzzy_refine,
              },
            },
          })
          end,
					desc = "Live grep",
				},
				{
					"<leader>fv",
          function()
            local lga_shortcuts = require("telescope-live-grep-args.shortcuts")

            lga_shortcuts.grep_visual_selection({
              postfix = " --iglob *",
            })
          end,
          mode = "v",
					desc = "Live grep visual selection",
				},
				{
					"<leader>bb",
          function()
            local builtin = require("telescope.builtin")
            local action_state = require("telescope.actions.state")

            builtin.buffers({
              initial_mode = "normal",
              attach_mappings = function(prompt_bufnr, map)
                local delete_buf = function()
                  local current_picker = action_state.get_current_picker(prompt_bufnr)
                  current_picker:delete_selection(function(selection)
                    vim.api.nvim_buf_delete(selection.bufnr, { force = true })
                  end)
                end

                map('n', '<c-d>', delete_buf)

                return true
              end
            }, {
              show_all_buffers = true,
              sort_mru = true,
            })
          end,
          desc = "Buffers",
				},
				{
					"<leader>fc",
					"<CMD>Telescope commands<CR>",
					desc = "Commands",
				},
				{
					"<leader>fh",
					"<CMD>Telescope help_tags<CR>",
					desc = "Help tags",
				},
				{
					"<leader>fo",
					"<CMD>Telescope oldfiles<CR>",
					desc = "Recent files",
				},
				{
					"<leader>ft",
					"<CMD>Telescope filetypes<CR>",
					desc = "Filetypes",
				},
				{
					"<leader>fF",
					"<CMD>Telescope<CR>",
					desc = "Telescope",
				},
				{
					"<leader>fu",
					"<CMD>Telescope undo<CR>",
					desc = "Undo history",
				},
			}
		end,
	},
}

return M

