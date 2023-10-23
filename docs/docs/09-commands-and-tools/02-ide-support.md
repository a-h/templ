# IDE support

## Visual Studio Code

There's a VS Code extension, just make sure you've already installed templ and that it's on your path.

- <https://marketplace.visualstudio.com/items?itemName=a-h.templ>
- <https://github.com/a-h/templ-vscode>

VSCodium users can find the extension on the Open VSX Registry at <https://open-vsx.org/extension/a-h/templ>

### Format on Save

Include the following into your settings.json to activate formatting `.templ` files on save with the
templ plugin:

```json
{
    "editor.formatOnSave": true,
    "[templ]": {
        "editor.defaultFormatter": "a-h.templ"
    },
{
```

### Tailwind CSS Intellisense

Include the following to the settings.json in order to enable autocompletion for Tailwind CSS in `.templ` files:

```json
{
  "tailwindCSS.includeLanguages": {
    "templ": "html"
  }
}
```

## Neovim &gt; 0.5.0

A vim / neovim plugin is available from <https://github.com/Joe-Davidson1802/templ.vim> which adds syntax highlighting.

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

In Neovim, you can use the `:LspInfo` command to check which Language Servers (if any) are running. If the expected language server has not started, it could be due to the unregistered templ file extension. To resolve this issue, add the following code to your configuration:

```lua
-- additional filetypes
vim.filetype.add({
 extension = {
  templ = "templ",
 },
})
```

### Format on Save

With the templ LSP installed and configured, you can use the following code snippet to format on save:


```lua
-- Format current buffer using LSP.
vim.api.nvim_create_autocmd(
  {
    -- 'BufWritePre' event triggers just before a buffer is written to file.
    "BufWritePre"
  },
  {
    pattern = {"*.templ"},
    callback = function()
      -- Format the current buffer using Neovim's built-in LSP (Language Server Protocol).
      vim.lsp.buf.format()
    end,
  }
)
```

### Tailwind CSS Intellisense

In order to enable autocompletion for Tailwindcss CSS in `.templ` files make sure to add the following config:

```lua
require("lspconfig").tailwindcss.setup({
  filetypes = {
    'templ'
    -- include any other filetypes where you need tailwindcss
  },
  init_options = {
    userLanguages = {
        templ = "html"
    }
  }
})
```

## Helix

https://helix-editor.com/

Helix has built-in templ support in unstable since https://github.com/helix-editor/helix/pull/8540/commits/084628d3e0c29f4021f53b3e45997ae92033d2d2

It will be included in official releases after version 23.05.

## Troubleshooting

### Check that go, gopls and templ are installed and are present in the path

```shell
which go gopls templ
```

You should see 3 lines returned, showing the location of each binary:

```
/run/current-system/sw/bin/go
/Users/adrian/go/bin/gopls
/Users/adrian/bin/templ
```

### Check that you can run the templ binary

Run `templ lsp --help`, you should see help text.

* If you can't run the `templ` command at the command line:
  * Check that the `templ` binary is within a directory that's in your path (`echo $PATH` for Linux/Mac/WSL, `$env:path` for Powershell).
  * Update your profile to ensure that the change to your path applies to new shells and processes.
    * On MacOS / Linux, you may need to update your `~/.zsh_profile`, `~/.bash_profile` or `~/.profile` file.
    * On Windows, you will need to use the "Environment Variables" dialog. For WSL, use the Linux config.
  * On MacOS / Linux, check that the file is executable and resolve it with `chmod +x /path/to/templ`.
  * On MacOS, you might need to go through the steps at https://support.apple.com/en-gb/guide/mac-help/mh40616/mac to enable binaries from an "unidentified developer" to run.
* If you're running VS Code using Windows Subsystem for Linux (WSL), then templ must also be installed within the WSL environment, not just inside your Windows environment.
* If you're running VS Code in a Devcontainer, it must be installed in there.


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
