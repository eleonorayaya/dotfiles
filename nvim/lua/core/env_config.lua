local go_enabled = vim.env.EQT_ENABLE_GO ~= nil
local python_enabled = vim.env.EQT_ENABLE_PYTHON ~= nil
local ruby_enabled = vim.env.EQT_ENABLE_RUBY ~= nil
local web_enabled = vim.env.EQT_ENABLE_WEB ~= nil

vim.g.btl_config = {
	go_enabled = go_enabled,
	python_enabled = python_enabled,
	ruby_enabled = ruby_enabled,
	web_enabled = web_enabled,
}
