package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

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
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		log.Fatalf("failed to parse git output: %v", err)
	}

	// The current commit isn't the one we're about to commit.
	fmt.Printf("0.2.%d", count+1)
}
