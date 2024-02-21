# GitHub Action for `templ`

For teams looking to automate their Go code generation from `.templ` files, there is a GitHub Action for that: [templ-generator-action](https://github.com/marketplace/actions/templ-generator-action). This action integrates seamlessly with your CI/CD pipeline, ensuring that your Go code is always synchronized with your templates.

## Features

- **Automatic Code Generation**: Automatically converts `.templ` files into Go source code with every push, keeping your codebase up-to-date.
- **Customizable Workflow**: Configure the action to fit your project's needs, with options for directory paths, commit messages, and more.
- **Easy Integration**: Add the action to your GitHub workflow with just a few lines of YAML.

## How to Use

To add the `templ-generator-action` to your workflow, just include it as a step in your GitHub Actions workflow file:

```yaml
- name: Generate templ code
  uses: capthiron/templ-generator-action@v1
```

For detailed usage and configuration options, please refer to the [GitHub repository](https://github.com/capthiron/templ-generator-action).
