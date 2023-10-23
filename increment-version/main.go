package main

import (
	"fmt"
	"log"
	"os/exec"
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
	count := strings.TrimSpace(string(output))

	fmt.Printf("0.2.%s", count)
}
