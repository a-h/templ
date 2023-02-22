module github.com/a-h/templ

go 1.19

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/a-h/lexical v0.0.53
	github.com/a-h/pathvars v0.0.0-20200320143331-78b263b728e2
	github.com/google/go-cmp v0.5.9
	github.com/hashicorp/go-multierror v1.1.1
	github.com/natefinch/atomic v1.0.1
	github.com/rs/cors v1.8.2
	go.lsp.dev/jsonrpc2 v0.10.0
	go.lsp.dev/protocol v0.12.1-0.20220402221718-d6b9de0e0b4d
	go.lsp.dev/uri v0.3.0
	go.uber.org/zap v1.21.0
	golang.org/x/mod v0.5.1
)

require (
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.3.4 // indirect
	go.lsp.dev/pkg v0.0.0-20210717090340-384b27a52fb2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8 // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
)

//replace github.com/a-h/lexical => /Users/adrian/github.com/a-h/lexical

replace go.lsp.dev/protocol => github.com/a-h/protocol v0.0.0-20230222141054-e5a15864b7f1
