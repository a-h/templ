---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**Before you begin**
Please make sure you're using the latest version of the templ CLI (`go install github.com/a-h/templ/cmd/templ@latest`), and have upgraded your project to use the latest version of the templ runtime (`go get -u github.com/a-h/templ@latest`)

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
A small, self-container, complete reproduction, uploaded to a Github repo, containing the minimum amount of files required to reproduce the behaviour, along with a list of commands that need to be run. Keep it simple.

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots or screen captures to help explain your problem.

**Logs**
If the issue is related to IDE support, run through the LSP troubleshooting section at https://templ.guide/commands-and-tools/ide-support/#troubleshooting-1 and include logs from templ

**`templ info` output**
Run `templ info` and include the output.

**Desktop (please complete the following information):**
 - OS: [e.g. MacOS, Linux, Windows, WSL]
 - templ CLI version (`templ version`)
- Go version (`go version`)
- `gopls` version (`gopls version`)

**Additional context**
Add any other context about the problem here.
