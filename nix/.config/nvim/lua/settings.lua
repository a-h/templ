vim.opt.autoindent = true
vim.opt.encoding = "utf-8"
vim.opt.fileencoding = "utf-8"
vim.opt.tabstop = 2

-- Set termguicolors to enable highlight groups.
vim.opt.termguicolors = true

-- Move the preview screen.
vim.opt.splitbelow = true

-- Make it so that the gutter (left column) doesn't move.
vim.opt.signcolumn = "yes"

-- Set line numbers to be visible all of the time.
vim.opt.number = true

-- Disable mouse control.
vim.cmd("set mouse=")

-- Use system clipboard.
--vim.api.nvim_command('set clipboard+=unnamedplus')
