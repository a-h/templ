-- Autoformat Go files on save and add goimports style fix-up.
-- See https://github.com/neovim/nvim-lspconfig/issues/115

vim.api.nvim_create_autocmd('BufWritePre', {
  pattern = '*.go',
  callback = function()
    vim.lsp.buf.code_action({ context = { only = { 'source.organizeImports' } }, apply = true })
  end
})
