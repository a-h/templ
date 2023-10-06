# IDE support

## vscode

There's a VS Code extension, just make sure you've already installed templ and that it's on your path. 

* https://marketplace.visualstudio.com/items?itemName=a-h.templ
* https://github.com/a-h/templ-vscode

## Neovim &gt; 0.5.0

A vim / neovim plugin is available from https://github.com/Joe-Davidson1802/templ.vim which adds syntax highlighting.

For neovim you can also use [nvim-treesitter](https://github.com/nvim-treesitter/nvim-treesitter) for syntax highlighting with the custom parser [tree-sitter-templ](https://github.com/vrischmann/tree-sitter-templ).

The configuration for the templ Language Server is included in [lspconfig](https://github.com/neovim/nvim-lspconfig), [mason](https://github.com/williamboman/mason.nvim),
and [mason-lspconfig](https://github.com/williamboman/mason-lspconfig.nvim).

Installing and configuring the templ LSP is no different to setting up any other Language Server.

```lua
-- Use a loop to conveniently call 'setup' on multiple servers and
-- map buffer local keybindings when the language server attaches
local servers = { 'gopls', 'ccls', 'cmake', 'tsserver', 'templ' }
for _, lsp in ipairs(servers) do
  nvim_lsp[lsp].setup {
    on_attach = on_attach,
    flags = {
      debounce_text_changes = 150,
    },
  }
end
```

The `templ` command must be in your system path for the LSP to be able to start.

If the language server fails to start, it could be due to the unregistered templ file extension. To resolve this issue, add the following code to your configuration: 

```lua
-- additional filetypes
vim.filetype.add({
	extension = {
		templ = "templ",
	},
})
```
