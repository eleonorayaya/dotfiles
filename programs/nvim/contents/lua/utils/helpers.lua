function _G.put(...)
  local objects = {}
  for i = 1, select("#", ...) do
    local v = select(i, ...)
    table.insert(objects, vim.inspect(v))
  end

  print(table.concat(objects, "\n"))
  return ...
end

function table.merge(t1, t2)
  for k, v in ipairs(t2) do
    table.insert(t1, v)
  end

  return t1
end

function _G.config_paths()
  put(vim.inspect(vim.api.nvim_list_runtime_paths()))
end

function _G.is_git_repo()
  vim.fn.system("git rev-parse --is-inside-work-tree")
  return vim.v.shell_error == 0
end

function _G.get_git_root()
  local dot_git_path = vim.fn.finddir(".git", ".;")
  return vim.fn.fnamemodify(dot_git_path, ":h")
end

function _G.get_current_branch()
  if not is_git_repo() then
    return ""
  end

  local out = vim.fn.systemlist("git rev-parse --abbrev-ref HEAD")

  if vim.v.shell_error ~= 0 then
    print(string.format("git failed with: %s", table.concat(out, "\n")))
    return ""
  end

  return out[1]
end

