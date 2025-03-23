#!/bin/bash
echo If
go test -fuzz=FuzzIf -fuzztime=120s
echo For
go test -fuzz=FuzzFor -fuzztime=120s
echo Switch
go test -fuzz=FuzzSwitch -fuzztime=120s
echo Case
go test -fuzz=FuzzCaseStandard -fuzztime=120s
echo Default
go test -fuzz=FuzzCaseDefault -fuzztime=120s
echo TemplExpression
go test -fuzz=FuzzTemplExpression -fuzztime=120s
echo Expression
go test -fuzz=FuzzExpression -fuzztime=120s
echo SliceArgs
go test -fuzz=FuzzSliceArgs -fuzztime=120s
echo Funcs
go test -fuzz=FuzzFuncs -fuzztime=120s
