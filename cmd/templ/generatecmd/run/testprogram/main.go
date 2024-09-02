package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// This is a test program. It is used only to test the behaviour of the run package.
// The run package is supposed to be able to run and stop programs. Those programs may start
// child processes, which should also be stopped when the parent program is stopped.

// For example, running `go run .` will compile an executable and run it.

// So, this program does nothing. It just waits for a signal to stop.

// In "Well behaved" mode, the program will stop when it receives a signal.
// In "Badly behaved" mode, the program will ignore the signal and continue running.

// The run package should be able to stop the program in both cases.

var badlyBehavedFlag = flag.Bool("badly-behaved", false, "If set, the program will ignore the stop signal and continue running.")

func main() {
	flag.Parse()

	mode := "Well behaved"
	if *badlyBehavedFlag {
		mode = "Badly behaved"
	}
	fmt.Printf("%s process %d started.\n", mode, os.Getpid())

	// Start a web server on a known port so that we can check that this process is
	// not running, when it's been started as a child process, and we don't know
	// its pid.
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%d", os.Getpid())
		})
		err := http.ListenAndServe("127.0.0.1:7777", nil)
		if err != nil {
			fmt.Printf("Error running web server: %v\n", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	if !*badlyBehavedFlag {
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	}
	for {
		select {
		case <-sigs:
			fmt.Printf("Process %d received signal. Stopping.\n", os.Getpid())
			return
		case <-time.After(1 * time.Second):
			fmt.Printf("Process %d still running...\n", os.Getpid())
		}
	}
}
