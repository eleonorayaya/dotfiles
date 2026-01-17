local M = {}

M.plugins = {
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    branch = "main",
    lazy = false,
    opts = {},
    config = function(_, opts)
      require("nvim-treesitter").setup(opts)

      -- Install parsers
      local parsers_to_install = {
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
      }

      -- Install parsers asynchronously
      require("nvim-treesitter").install(parsers_to_install)

      -- Enable treesitter highlighting for all filetypes
      vim.api.nvim_create_autocmd("FileType", {
        pattern = "*",
        callback = function()
          pcall(vim.treesitter.start)
        end,
      })

      vim.filetype.add({
        extension = {
          gotmpl = "gotmpl",
          ["yaml.tmpl"] = "yaml",
          ["yaml.tpl"] = "yaml",
          ["yml.tmpl"] = "yaml",
          ["yml.tpl"] = "yaml",
        },
        pattern = {
          [".*/templates/.*%.tpl"] = "helm",
          [".*/templates/.*%.ya?ml"] = "helm",
          ["helmfile.*%.ya?ml"] = "helm",
          -- Crossplane composition patterns
          [".*composition.*%.ya?ml"] = "yaml",
          [".*function.*%.ya?ml"] = "yaml",
          -- Files containing Crossplane API versions
          [".*crossplane%.io.*%.ya?ml"] = "yaml",
          [".*fn%.crossplane%.io.*%.ya?ml"] = "yaml",
          -- Kustomize patterns
          ["kustomization%.ya?ml"] = "yaml",
          [".*kustomize.*%.ya?ml"] = "yaml",
          -- Helm values files
          ["values.*%.ya?ml"] = "yaml",
          [".*values%.ya?ml"] = "yaml",
          ["Chart%.ya?ml"] = "yaml",
          -- Additional patterns for files with Go templating
          [".*%.gotmpl%.ya?ml"] = "yaml",
          [".*%.tpl%.ya?ml"] = "yaml",
        },
      })
    end,
  },
}

return M

