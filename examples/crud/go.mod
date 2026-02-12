module github.com/a-h/templ/examples/crud

go 1.25.0

toolchain go1.25.7

require (
	github.com/a-h/kv v0.0.0-20251001131013-326dbe4b4060
	github.com/a-h/templ v0.3.924
	github.com/gorilla/schema v1.4.1
	github.com/segmentio/ksuid v1.0.4
	zombiezen.com/go/sqlite v1.4.2
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93 // indirect
	golang.org/x/sys v0.39.0 // indirect
	modernc.org/libc v1.67.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.42.2 // indirect
)

replace github.com/a-h/templ => ../../
