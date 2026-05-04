return {
  {
    "romus204/tree-sitter-manager.nvim",
    lazy = false,
    config = function()
      require("tree-sitter-manager").setup({
        ensure_installed = {
          "bash",
          "c",
          "comment",
          "css",
          "csv",
          "cue",
          "diff",
          "dockerfile",
          "git_config",
          "git_rebase",
          "gitattributes",
          "gitcommit",
          "gitignore",
          "go",
          "gomod",
          "gosum",
          "gotmpl",
          "gowork",
          "hcl",
          "helm",
          "html",
          "java",
          "javascript",
          "jinja",
          "jinja_inline",
          "jq",
          "jsdoc",
          "json",
          "lua",
          "luadoc",
          "luap",
          "markdown",
          "markdown_inline",
          "nix",
          "nu",
          "python",
          "query",
          "regex",
          "ruby",
          "rust",
          "scss",
          "terraform",
          "toml",
          "tsv",
          "typescript",
          "vim",
          "vimdoc",
          "xml",
          "yaml",
        },
        auto_install = true,
        highlight = true,
        nohighlight = { "log", "txt", "csv", "json" },
      })

      vim.api.nvim_create_autocmd({ "BufReadPost", "BufEnter" }, {
        group = vim.api.nvim_create_augroup("TreesitterLargeFile", { clear = true }),
        callback = function(args)
          local buf = args.buf
          local bufname = vim.api.nvim_buf_get_name(buf)
          if bufname == "" or vim.fn.filereadable(bufname) ~= 1 then
            return
          end
          local buftype = vim.bo[buf].buftype
          if buftype ~= "" and buftype ~= "acwrite" then
            return
          end
          local line_count = vim.api.nvim_buf_line_count(buf)
          local file_size = vim.fn.getfsize(bufname)
          if line_count > 5000 or file_size > 2 * 1024 * 1024 then
            vim.treesitter.stop(buf)
          end
        end,
      })
    end,
  },
}
