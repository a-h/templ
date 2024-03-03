echo "Resetting git repo"
git reset --hard
echo "Updating to templ version" $TEMPL_VERSION
RUN go get github.com/a-h/templ@$TEMPL_VERSION
echo "Generating code" templ@$TEMPL_VERSION
templ generate
echo "Running tests"
go test ./...
go build ./...
echo $TEMPL_PREFIX "OK" templ@$TEMPL_VERSION
