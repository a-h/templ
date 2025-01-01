# IDE support

## Visual Studio Code

There's a VS Code extension, just make sure you've already installed templ and that it's on your path.

- https://marketplace.visualstudio.com/items?itemName=a-h.templ
- https://github.com/a-h/templ-vscode

VSCodium users can find the extension on the Open VSX Registry at https://open-vsx.org/extension/a-h/templ

### Format on Save

Include the following into your settings.json to activate formatting `.templ` files on save with the
templ plugin:

```json
{
    "editor.formatOnSave": true,
    "[templ]": {
        "editor.defaultFormatter": "a-h.templ"
    },
}
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
:::note 
Tailwind language servers require a tailwind.config.js file to be present in the root of your project. You can create a new config file with `npx tailwindcss init`, or use samples available at https://tailwindcss.com/docs/configuration
:::

### Emmet HTML completion

Include the following to the settings.json in order to get smooth HTML completion via emmet (such as expanding `input:button<Tab>` to `<input type="button" value="">`). The emmet plugin is built into vscode and just needs to be activated for `.templ` files:

```json
{
  "emmet.includeLanguages": {
    "templ": "html"
  }
}
```

## Neovim &gt; 0.5.0

A plugin written in VimScript which adds syntax highlighting: [joerdav/templ.vim](https://github.com/Joe-Davidson1802/templ.vim).

For neovim you can use [nvim-treesitter](https://github.com/nvim-treesitter/nvim-treesitter) and install [tree-sitter-templ](https://github.com/vrischmann/tree-sitter-templ) with `:TSInstall templ`.

The configuration for the templ Language Server is included in [lspconfig](https://github.com/neovim/nvim-lspconfig), [mason](https://github.com/williamboman/mason.nvim),
and [mason-lspconfig](https://github.com/williamboman/mason-lspconfig.nvim).

The `templ` command must be in your system path for the LSP to be able to start. Ensure that you can run it from the command line before continuing.

Installing and configuring the templ LSP is no different to setting up any other Language Server.

```lua
local lspconfig = require("lspconfig")

-- Use a loop to conveniently call 'setup' on multiple servers and
-- map buffer local keybindings when the language server attaches

local servers = { 'gopls', 'ccls', 'cmake', 'tsserver', 'templ' }
for _, lsp in ipairs(servers) do
  lspconfig[lsp].setup({
    on_attach = on_attach,
    capabilities = capabilities,
  })
end
```

In Neovim, you can use the `:LspInfo` command to check which Language Servers (if any) are running. If the expected language server has not started, it could be due to the unregistered templ file extension. 

To resolve this issue, add the following code to your configuration. This is also necessary for other LSPs to "pick up" on .templ files.

```lua
vim.filetype.add({ extension = { templ = "templ" } })
```

##### Other LSPs within .templ files

These LSPs can be used *in conjunction* with the templ lsp and tree-sitter. Here's how to set them up.

[html-lsp](https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md#html) - First make sure you have it installed `:LspInstall html` or find it on the `:Mason` list. 

```lua
lspconfig.html.setup({
    on_attach = on_attach,
    capabilities = capabilities,
    filetypes = { "html", "templ" },
})
```

[htmx-lsp](https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md#htmx) - First make sure you have it installed `:LspInstall htmx` or find it on the `:Mason` list. Note with this LSP, it activates after you type `hx-` in an html attribute, because that's how all htmx attributes are written.

```lua
lspconfig.htmx.setup({
    on_attach = on_attach,
    capabilities = capabilities,
    filetypes = { "html", "templ" },
})
```

[tailwindcss](https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md#tailwindcss) - First make sure you have it installed `:LspInstall tailwindcss` or find it on the `:Mason` list.

```lua
lspconfig.tailwindcss.setup({
    on_attach = on_attach,
    capabilities = capabilities,
    filetypes = { "templ", "astro", "javascript", "typescript", "react" },
    settings = {
      tailwindCSS = {
        includeLanguages = {
          templ = "html",
        },
      },
    },
})
```

Inside of your `tailwind.config.js`, you need to tell tailwind to look inside of .templ files and/or .go files.

:::tip
If you don't have a `tailwind.config.js` in the root directory of your project, the Tailwind LSP won't activate, and you won't see autocompletion results.
:::

```js
module.exports = {
    content: [ "./**/*.html", "./**/*.templ", "./**/*.go", ],
    theme: { extend: {}, },
    plugins: [],
}
```

### Formatting

With the templ LSP installed and configured, you can use the following code snippet to format on save:


```lua
vim.api.nvim_create_autocmd({ "BufWritePre" }, { pattern = { "*.templ" }, callback = vim.lsp.buf.format })
```
`BufWritePre` means that the callback gets ran after you call `:write`.

If you have multiple LSPs attached to the same buffer, and you have issues with `vim.lsp.buf.format`, you can use this snippet to run `templ fmt` in the same way that you might from the command line.

This will get the buffer and its corresponding filename, and refresh the buffer after it has been formatted so you don't get out of sync issues.

```lua
local custom_format = function()
    if vim.bo.filetype == "templ" then
        local bufnr = vim.api.nvim_get_current_buf()
        local filename = vim.api.nvim_buf_get_name(bufnr)
        local cmd = "templ fmt " .. vim.fn.shellescape(filename)

        vim.fn.jobstart(cmd, {
            on_exit = function()
                -- Reload the buffer only if it's still the current buffer
                if vim.api.nvim_get_current_buf() == bufnr then
                    vim.cmd('e!')
                end
            end,
        })
    else
        vim.lsp.buf.format()
    end
end
```

To apply this `custom_format` in your neovim configuration as a keybinding, apply it to the `on_attach` function.

```lua
local on_attach = function(client, bufnr)
    local opts = { buffer = bufnr, remap = false }
    -- other configuration options
    vim.keymap.set("n", "<leader>lf", custom_format, opts)
end
```

To make this `custom_format` run on save, make the same autocmd from before and replace the callback with `custom_format`. 

```lua
vim.api.nvim_create_autocmd({ "BufWritePre" }, { pattern = { "*.templ" }, callback = custom_format })
```

You can also rewrite the function like so, given that the function will only be executed on .templ files.

```lua
local templ_format = function()
    local bufnr = vim.api.nvim_get_current_buf()
    local filename = vim.api.nvim_buf_get_name(bufnr)
    local cmd = "templ fmt " .. vim.fn.shellescape(filename)

    vim.fn.jobstart(cmd, {
        on_exit = function()
            -- Reload the buffer only if it's still the current buffer
            if vim.api.nvim_get_current_buf() == bufnr then
                vim.cmd('e!')
            end
        end,
    })
end
```

### Troubleshooting

If you cannot run `:TSInstall templ`, ensure you have an up-to-date version of [tree-sitter](https://github.com/nvim-treesitter/nvim-treesitter). The [package for templ](https://github.com/vrischmann/tree-sitter-templ) was [added to the main tree-sitter repositry](https://github.com/nvim-treesitter/nvim-treesitter/pull/5667) so you shouldn't need to install a separate plugin for it.

If you still don't get syntax highlighting after it's installed, try running `:TSBufEnable highlight`. If you find that you need to do this every time you open a .templ file, you can run this autocmd to do it for your neovim configuation.

```lua
vim.api.nvim_create_autocmd("BufEnter", { pattern = "*.templ", callback = function() vim.cmd("TSBufEnable highlight") end }) 
```

### Minimal Config

Minimal config with the following features (useful for debugging):
- [lazy-vim](https://github.com/folke/lazy.nvim): neovim package manager
- [lsp config](https://github.com/neovim/nvim-lspconfig)
    - templ-lsp
    - html-lsp
    - htmx-lsp
    - tailwind-lsp
- [cmp](https://github.com/hrsh7th/nvim-cmp): for autocompletion
- [tree-sitter](https://github.com/nvim-treesitter/nvim-treesitter): for synx highlighting
    - [tree sitter templ](https://github.com/vrischmann/tree-sitter-templ)

To use this config:

* As a permanent setup: Create/replace `init.lua` in your config folder (`~/.config/nvim/`)
* As a temporary setup: create a new folder in your `.config` (e.g.`~/.config/nvim_test`) and tell neovim to start up with that as the nvim appname `NVIM_APPNAME=nvim_test nvim` (see [neovim docs](https://neovim.io/doc/user/starting.html#%24NVIM_APPNAME) for further explanation.


```lua
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git",
    "clone",
    "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable", -- latest stable release
    lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)
vim.g.mapleader = " " -- Make sure to set `mapleader` before lazy so your mappings are correct

require("lazy").setup({
  'neovim/nvim-lspconfig',
  {
    -- Autocompletion
    'hrsh7th/nvim-cmp',
    dependencies = {
      'hrsh7th/cmp-nvim-lsp',
    },
  },
  {
    -- Highlight, edit, and navigate code
    'nvim-treesitter/nvim-treesitter',
    dependencies = {
      'vrischmann/tree-sitter-templ',
    },
    build = ':TSUpdate',
  },
})

vim.filetype.add({ extension = { templ = "templ" } })

capabilities = require('cmp_nvim_lsp').default_capabilities(capabilities)
local lspconfig = require("lspconfig")

lspconfig.templ.setup{
  on_attach = on_attach,
  capabilities = capabilities,
}

lspconfig.tailwindcss.setup({
  on_attach = on_attach,
  capabilities = capabilities,
  filetypes = { "templ", "astro", "javascript", "typescript", "react" },
  init_options = { userLanguages = { templ = "html" } },
})

lspconfig.html.setup({
  on_attach = on_attach,
  capabilities = capabilities,
  filetypes = { "html", "templ" },
})

lspconfig.htmx.setup({
  on_attach = on_attach,
  capabilities = capabilities,
  filetypes = { "html", "templ" },
})

local cmp = require 'cmp'
cmp.setup({
  mapping = cmp.mapping.preset.insert({
    ['<C-b>'] = cmp.mapping.scroll_docs(-4),
    ['<C-f>'] = cmp.mapping.scroll_docs(4),
    ['<C-Space>'] = cmp.mapping.complete(),
    ['<C-e>'] = cmp.mapping.abort(),
    ['<CR>'] = cmp.mapping.confirm({ select = true }),
  }),
  sources = cmp.config.sources({
    { name = 'nvim_lsp' },
  })
})

require'nvim-treesitter.configs'.setup {
  ensure_installed = { "templ" },
  sync_install = false,
  auto_install = true,
  ignore_install = { "javascript" },
  highlight = {
    enable = true,
  },
}
```

## Vim

Currently support for vim is limited. Configure formatting with the following VimScript:

```vim
set autoread
autocmd BufWritePost *.templ silent! execute "!PATH=\"$PATH:$(go env GOPATH)/bin\" templ fmt <afile> >/dev/null 2>&1" | redraw!
```

## Helix

https://helix-editor.com/

Helix has built-in templ support in unstable since https://github.com/helix-editor/helix/pull/8540/commits/084628d3e0c29f4021f53b3e45997ae92033d2d2

It will be included in official releases after version 23.05.

## Emacs

[templ-ts-mode](https://github.com/danderson/templ-ts-mode) is a major mode for templ files that provides syntax highlighting, indentation, and the other usual major mode things. It is available on [MELPA](https://melpa.org/#/templ-ts-mode) and can be installed like any other Emacs package.

Templ support requires the [tree-sitter parser for Templ](https://github.com/vrischmann/tree-sitter-templ). If the parser is missing, the mode asks you on first use whether you want to download and build it via `treesit-install-language-grammar` (requires git and a C compiler).

## Troubleshooting

### Check that go, gopls and templ are installed and are present in the path

```bash
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
    "templ.log": "/Users/adrian/templ.log",
    "templ.goplsLog": "/Users/adrian/gopls.log",
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

For IntelliJ, configure the plugin settings `.idea/templ.xml`.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="TemplSettings">
    <option name="goplsLog" value="$USER_HOME$/gopls.log" />
    <option name="goplsRPCTrace" value="true" />
    <option name="http" value="localhost:7575" />
    <option name="log" value="$USER_HOME$/templ.log" />
  </component>
</project>
```

### Make a minimal reproduction, and include the logs

The logs can be quite verbose, since almost every keypress results in additional logging. If you're thinking about submitting an issue, please try and make a minimal reproduction.

### Look at the web server

The web server option provides an insight into the internal state of the language server. It may provide insight into what's going wrong.

### Run templ info

The `templ info` command outputs information that's useful for debugging issues.
