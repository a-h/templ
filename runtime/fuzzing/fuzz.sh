#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"
FUZZTIME="${FUZZTIME:-120s}"
for target in FuzzScriptStringRoundTrip FuzzComponentAny FuzzAttributeName FuzzAttributeValue; do
	echo "$target"
	go test -run '^$' -fuzz "^${target}\$" -fuzztime "$FUZZTIME" .
done
