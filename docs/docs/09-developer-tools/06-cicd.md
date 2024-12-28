# Ensuring templ files have been committed

It's common practice to commit generated `*_templ.go` files to your source code repository, so that your codebase is always in a state where it can be built and run without needing to run `templ generate`, e.g. by running `go install` on your project, or by importing it as a dependency in another project.

In your CI/CD pipeline, if you want to check that `templ generate` has been ran on all templ files (with the same version of templ used by the CI/CD pipeline), you can run `templ generate` again.

If any files have changed, then the pipeline should fail, as this would indicate that the generated files are not up-to-date with the templ files.

```bash
templ generate
git diff --exit-code
```
