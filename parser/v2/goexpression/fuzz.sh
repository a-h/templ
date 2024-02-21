#echo If
#go test -fuzz=FuzzIf -fuzztime=60s
#echo For
#go test -fuzz=FuzzFor -fuzztime=60s
#echo Switch
#go test -fuzz=FuzzSwitch -fuzztime=60s
#echo Case
#go test -fuzz=FuzzCase -fuzztime=60s
echo Default
go test -fuzz=FuzzCaseDefault -fuzztime=60s
echo Expression
go test -fuzz=FuzzExpression -fuzztime=60s
echo SliceArgs
go test -fuzz=FuzzSliceArgs -fuzztime=60s
