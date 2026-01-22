local M = {}

M.plugins = {
  {
    "nvim-neo-tree/neo-tree.nvim",
    branch = "v3.x",
    event = "BufReadPre",
    dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-tree/nvim-web-devicons",
      "MunifTanjim/nui.nvim",
      "folke/snacks.nvim",
    },
    config = function()
      vim.g.neo_tree_remove_legacy_commands = 1

      require("neo-tree").setup({
        event_handlers = {
          {
            event = "neo_tree_buffer_enter",
            handler = function(arg)
              vim.opt_local.statuscolumn = "%s"
              vim.opt_local.number = false
              vim.opt_local.relativenumber = false
              vim.opt_local.wrap = true
              vim.opt_local.sidescrolloff = 0
            end,
          },
        },
        enable_modified_markers = false,
        enable_opened_markers = false,
        close_if_last_window = true,
        enable_diagnostics = true,
        sort_case_insensitive = true,
        default_component_configs = {
          indent = {
            with_expanders = true,
            with_markers = false,
          },
          icon = {
            folder_closed = "",
            folder_open = "",
          },
          name = {
            use_git_status_colors = true,
          },
          git_status = {
            symbols = {
              added = "",
              deleted = "",
              modified = "",
              renamed = "",
              untracked = "",
              ignored = "",
              unstaged = "",
              staged = "",
              conflict = "",
            },
          },
        },
        window = {
          mappings = {
            ["<2-LeftMouse>"] = "open",
            ["<CR>"] = "open",
            ["<esc>"] = "revert_preview",
            ["P"] = {
              "toggle_preview",
              config = {
                use_float = true,
              },
            },
            ["s"] = "open_split",
            ["v"] = "open_vsplit",
            ["w"] = "open_with_window_picker",
            ["z"] = "close_all_nodes",
            ["Z"] = "expand_all_nodes",
            ["a"] = {
              "add",
              config = {
                show_path = "none",
              },
            },
            ["d"] = "delete",
            ["r"] = "rename",
            ["y"] = "copy_to_clipboard",
            ["x"] = "cut_to_clipboard",
            ["p"] = "paste_from_clipboard",
            ["c"] = "copy",
            ["m"] = "move",
            ["q"] = "close_window",
            ["?"] = "show_help",
            ["<"] = "prev_source",
            [">"] = "next_source",
          },
        },
        filesystem = {
          commands = {
          },
          filtered_items = {
            hide_dotfiles = false,
            hide_gitignored = false,
            never_show = {
              ".DS_Store",
            },
            visible = true,
          },
          follow_current_file = {
            enabled = true,
          },
          use_libuv_file_watcher = true,
          event_handlers = function()
            local function on_move(data)
              require("snacks").rename.on_rename_file(data.source, data.destination)
            end
            return {
              {
                event = require("neo-tree.events").FILE_MOVED,
                handler = on_move,
              },
              {
                event = require("neo-tree.events").FILE_RENAMED,
                handler = on_move,
              },
            }
          end,
          window = {
          },
          find_command = "fd",
          find_args = {
            fd = {
              "--hidden",
              "--exclude",
              ".git",
              "--exclude",
              "node_modules",
            },
          },
        },
      })

      vim.api.nvim_set_hl(0, "NeoTreeFileName_35", { fg = "#606377", bg = nil, bold = true })
      vim.api.nvim_set_hl(0, "NeoTreeRootName_35", { fg = "#606377", bg = nil, bold = true })
      vim.api.nvim_set_hl(0, "NeoTreeMessage", { fg = "#606377", bg = nil, bold = false })
    end,
    keys = {
      {
        "<D-b>",
        function()
          vim.cmd("Neotree toggle")
          -- TODO: what is this?
          vim.cmd("wincmd p")
        end,
        desc = "Toggle file tree",
        mode = { "n", "t" },
      },
      {
        "<leader>of",
        function()
          vim.cmd("Neotree toggle")
        end,
        desc = "Reveal open file in tree",
      },
    },
  },
}

return M

