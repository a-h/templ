package runtime

import (
	"os"

	"github.com/a-h/templ/watchstrings"
)

var WatchMode bool = false

func init() {
	WatchMode = os.Getenv("TEMPL_WATCH_MODE") == "true"
}

func Watch(ptr *[]string) {
	watchstrings.Watch(ptr)
}
