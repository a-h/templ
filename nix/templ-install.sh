# Create the standard environment.
source $stdenv/setup
# Go requires a cache directory so let's point it at one.
mkdir -p /tmp/go-cache
export GOCACHE=$TMPDIR/go-cache
export GOMODCACHE=$TMPDIR/go-cache
# Build the source code.
cd $src/cmd/templ
# Build the templ binary and output it.
mkdir -p $out/bin
go build -o $out/bin/templ
