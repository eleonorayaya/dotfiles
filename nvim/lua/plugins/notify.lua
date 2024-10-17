return {
	"rcarriga/nvim-notify",
	config = function()
		local notify = require("notify")
		notify.setup({
			render = "compact",
		})
		vim.notify = notify
	end,
}
