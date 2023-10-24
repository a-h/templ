# Contributing to Templ

Thank you for considering contributing to Templ! Your contributions are vital to the project's success, and we're thrilled to have you onboard. Before you get started, please take a moment to review our contribution guidelines.

## Project Overview

**Templ** is an HTML templating language for Go, specifically designed with robust developer tooling in mind. To ensure that your contributions align with the project's goals, it's important to understand how the project works, its structure, and testing procedures.

### Testing

Templ relies on an extensive testing suite. It's crucial to comprehend how Templ is tested, including the evaluation of the parser and generation processes. Please note that all tests are named as assertions, and it is expected that new contributions follow this naming convention.

## Raising New Features

If you have an idea for a new feature, we strongly encourage you to follow these steps:

1. **Discuss the Feature Design**: Before diving into coding, initiate a discussion with the project maintainers. This step helps ensure that your feature aligns with the project's objectives and saves you from investing time in a design that may not work.

2. **Provide Examples**: When relevant, include examples of input Templ files and describe the expected output HTML. This helps the development team and other contributors understand your vision more clearly.

## Reporting Issues

If you encounter a bug or have an idea for an improvement, please follow these guidelines when reporting issues:

1. **Reproduction Repository**: Providing a reproduction repository for the issue is highly valuable. This helps the maintainers understand the problem and work towards a solution effectively.

2. **Failing Unit Tests**: If possible, create a failing unit test that reproduces the issue you're facing. This provides a clear and automated way to describe the problem.

3. **LSP Troubleshooting**: When dealing with Language Server Protocol (LSP) issues, refer to the troubleshooting guide available in the online documentation for guidance.

## Additional Guidance

To contribute effectively to Templ, consider the following additional guidance:

- **Using the `xc` Task Runner**: Learn how to use the `xc` task runner to execute tasks, including running unit tests. Ideally, we encourage you to update the `flake.nix` to include a development shell for building and testing the software.

- **Conventional Commit Syntax**: The project follows a conventional commit syntax. Please reference this syntax when making commits.

- **Coding Style**: Maintain a clean and non-nested coding style. Ensure that comments are complete sentences and end with a full stop. Minimize unnecessary line breaks; use line breaks to separate "paragraphs" of code.

- **Linter**: There is a linter in the continuous integration (CI) process. It is recommended that you run the linter locally with `xc` before submitting a pull request (PR).

- **Generator Changes**: If you make changes to the generator, remember to run `xc generate` to regenerate all files. This is automatically handled by the `xc test` command.

## Code of Conduct

Please be respectful and considerate when interacting with other contributors. We have a [Code of Conduct](LICENSE) in place to foster a positive and inclusive environment for everyone.

Your contributions play a pivotal role in making Templ even better, and we look forward to your involvement. If you have any questions or need further clarification, feel free to reach out to the project maintainers.

Thank you for being part of the Templ community and for helping enhance this project!
