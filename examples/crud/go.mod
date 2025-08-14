module github.com/a-h/templ/examples/crud

go 1.23.4

toolchain go1.24.4

require (
	github.com/a-h/kv v0.0.0-20250429164327-e923751a9968
	github.com/a-h/templ v0.3.924
	github.com/gorilla/csrf v1.7.3
	github.com/gorilla/schema v1.4.1
	github.com/segmentio/ksuid v1.0.4
	zombiezen.com/go/sqlite v1.4.2
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	modernc.org/libc v1.65.7 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.37.1 // indirect
)

replace github.com/a-h/templ => ../../
