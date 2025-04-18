package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"golang.org/x/tools/go/packages"
)

func main() {
	args := append([]string{"run", "--", "@io_bazel_rules_go//go/tools/gopackagesdriver"}, os.Args[1:]...)
	cmd := exec.Command("bazel", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		os.Stderr.WriteString("bazel run failed: " + err.Error() + "\n")
		os.Exit(1)
	}

	var resp packages.DriverResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		os.Stderr.WriteString("decode failed: " + err.Error() + "\n")
		os.Exit(1)
	}

	for i, pkg := range resp.Packages {
		dirs := map[string]struct{}{}
		for _, f := range pkg.GoFiles {
			dirs[filepath.Dir(f)] = struct{}{}
		}
		for dir := range dirs {
			matches, _ := filepath.Glob(filepath.Join(dir, "*.templ"))
			for _, m := range matches {
				if !slices.Contains(pkg.OtherFiles, m) {
					resp.Packages[i].OtherFiles = append(resp.Packages[i].OtherFiles, m)
				}
			}
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		os.Stderr.WriteString("encode failed: " + err.Error() + "\n")
		os.Exit(1)
	}
}

