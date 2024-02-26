# Dependents tests

This project tests projects that use templ before deploying major changes to it.

## Tasks

### update-dependents

Get a list of all public projects in Github that use templ.

```bash
nix run a-h/github-download-dependents-info -- a-h/templ --csv dependents.csv
```

### clone-repos

Clone them all.

```bash
go run main.go -access-token=`pass github.com/read-public-repos`
```

### build-containers

Build `templ-dependent:previous` and `templ-dependent:current` Docker containers that contain the expected versions of the templ CLI.

```bash
docker build \
  --build-arg TEMPL_VERSION=v0.2.543 \
  -t templ-dependent:previous \
  .
docker build \
  --build-arg TEMPL_VERSION=ee2ba0e937dae19cf3bd1ee532ff3dcda5a8aae4 \
  -t templ-dependent:current \
  .
```

### test

```bash
docker run -e TEMPL_PREFIX="random_number" -v `pwd`/testdata/yokaracho/calculator-calories:/app templ-dependent:previous
docker run -e TEMPL_PREFIX="random_number" -v `pwd`/testdata/yokaracho/calculator-calories:/app templ-dependent:current
```
