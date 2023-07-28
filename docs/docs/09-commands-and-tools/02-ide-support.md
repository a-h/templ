# IDE support

## vscode

There's a VS Code extension, just make sure you've already installed templ and that it's on your path. 

* https://marketplace.visualstudio.com/items?itemName=a-h.templ
* https://github.com/a-h/templ-vscode

## Neovim &gt; 0.5.0

A vim / neovim plugin is available from https://github.com/Joe-Davidson1802/templ.vim which adds syntax highlighting.

For neovim you can also use [nvim-treesitter](https://github.com/nvim-treesitter/nvim-treesitter) for syntax highlighting with the custom parser [tree-sitter-templ](https://github.com/vrischmann/tree-sitter-templ).

To enable the built-in Language Server support of Neovim 5.x add the following code to your `.vimrc` prior to calling `setup` on the language servers, e.g.:

```lua
-- Add templ configuration.
local configs = require'lspconfig/configs'
if not nvim_lsp.templ then
  configs.templ = {
    default_config = {
      cmd = {"templ", "lsp"},
      filetypes = {'templ'},
      root_dir = nvim_lsp.util.root_pattern("go.mod", ".git"),
      settings = {},
    };
  }
end

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
