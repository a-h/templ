#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"
FUZZTIME="${FUZZTIME:-120s}"
for target in FuzzIf FuzzFor FuzzSwitch FuzzCaseStandard FuzzCaseDefault FuzzTemplExpression FuzzExpression FuzzSliceArgs FuzzFuncs; do
	echo "$target"
	go test -run '^$' -fuzz "^${target}\$" -fuzztime "$FUZZTIME" .
done
