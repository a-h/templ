# IDE support

## Visual Studio Code

There's a VS Code extension, just make sure you've already installed templ and that it's on your path. 

* https://marketplace.visualstudio.com/items?itemName=a-h.templ
* https://github.com/a-h/templ-vscode

VSCodium users can find the extension on the Open VSX Registry at https://open-vsx.org/extension/a-h/templ

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

## Troubleshooting

### Check that go, gopls and templ are installed and are present in the path.

```shell
which go gopls templ
```

You should see 3 lines returned, showing the location of each binary:

```
/run/current-system/sw/bin/go
/Users/adrian/go/bin/gopls
/Users/adrian/bin/templ
```

### Enable LSP logging

For VS Code, use the "Preferences: Open User Settings (JSON)" command in VS Code and add the configuration options.

```js
{
    // More settings...
    "templ.log": "/Users/adrian/logs/vscode-templ.txt",
    "templ.goplsLog": "/Users/adrian/logs/vscode-gopls.txt",
    "templ.http": "localhost:7575",
    "templ.goplsRPCTrace": true,
    "templ.pprof": false,
    // More stuff...
}
```

For Neovim, configure the LSP command to add the additional command line options.

```lua
local configs = require('lspconfig.configs')
configs.templ = {
  default_config = {
    cmd = { "templ", "lsp", "-http=localhost:7474", "-log=/Users/adrian/templ.log" },
    filetypes = { 'templ' },
    root_dir = nvim_lsp.util.root_pattern("go.mod", ".git"),
    settings = {},
  },
}
```

### Make a minimal reproduction, and include the logs

The logs can be quite verbose, since almost every keypress results in additional logging. If you're thinking about submitting an issue, please try and make a minimal reproduction.

### Look at the web server

The web server option provides an insight into the internal state of the language server. It may provide insight into what's going wrong.
