package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var fileFlag = flag.String("file", ".version", "Set the name of the file to modify")

func main() {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		log.Fatalf("failed to find git on path: %v", err)
	}

	cmd := exec.Command(gitPath, "rev-list", "main", "--count")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to run git: %v", err)
	}
	count := strings.TrimSpace(string(output))

	var dirty string
	cmd = exec.Command(gitPath, "diff", "--quiet")
	if cmd.Run() != nil {
		dirty = "-next"
	}

	fmt.Printf("0.2.%s%s", count, dirty)
}
