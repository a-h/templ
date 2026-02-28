package prettier

import (
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func Test(t *testing.T) {
	archive, err := txtar.ParseFile("testdata.txtar")
	if err != nil {
		t.Fatalf("failed to read testdata.txtar: %v", err)
	}
	for i := 0; i < len(archive.Files)-1; i += 2 {
		if archive.Files[i].Name != archive.Files[i+1].Name {
			t.Fatalf("test archive is not in the expected format: file pair at index %d do not match: %q vs %q", i, archive.Files[i].Name, archive.Files[i+1].Name)
		}
		t.Run(archive.Files[i].Name, func(t *testing.T) {
			inputData := archive.Files[i].Data
			expectedData := archive.Files[i+1].Data
			input := strings.TrimSpace(string(inputData))
			expected := strings.TrimSpace(string(expectedData))
			actual, err := Run(input, archive.Files[i].Name, DefaultCommand())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(actual) != expected {
				t.Errorf("Actual:\n%s\nExpected:\n%s", actual, expected)
			}
		})
	}
}

func TestIsAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping IsAvailable test in short mode")
	}
	t.Run("non-existent commands return false", func(t *testing.T) {
		var nonExistentCommand = "templ_non_existent_command --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME"
		if IsAvailable(nonExistentCommand) {
			t.Errorf("IsAvailable should return false for non-existent command %q", nonExistentCommand)
		}
	})
	t.Run("existing commands return true", func(t *testing.T) {
		if !IsAvailable("ls -lah") {
			t.Errorf("IsAvailable should return true for existing command %q", DefaultCommand)
		}
	})
}

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		shell    string
		command  string
		wantPath string
		wantArgs []string
	}{
		{
			name:     "nushell uses a custom command",
			goos:     "linux",
			shell:    "/usr/bin/nu",
			command:  shellNameToCommand["nu"],
			wantPath: "/usr/bin/nu",
			wantArgs: []string{"/usr/bin/nu", "-c", shellNameToCommand["nu"]},
		},
		{
			name:     "bash uses the default posix command",
			goos:     "linux",
			shell:    "/bin/bash",
			command:  defaultPosixCommand,
			wantPath: "/bin/bash",
			wantArgs: []string{"/bin/bash", "-c", defaultPosixCommand},
		},
		{
			name:     "zsh uses the default posix command",
			goos:     "linux",
			shell:    "/bin/zsh",
			command:  defaultPosixCommand,
			wantPath: "/bin/zsh",
			wantArgs: []string{"/bin/zsh", "-c", defaultPosixCommand},
		},
		{
			name:     "empty shell defaults to the default posix command",
			goos:     "linux",
			shell:    "",
			command:  defaultPosixCommand,
			wantPath: "/bin/sh",
			wantArgs: []string{"/bin/sh", "-c", defaultPosixCommand},
		},
		{
			name:     "windows uses cmd.exe regardless of shell",
			goos:     "windows",
			shell:    "C:\\Program Files\\PowerShell\\pwsh.exe",
			command:  "prettier --stdin-filepath test.html",
			wantPath: "cmd.exe",
			wantArgs: []string{"cmd.exe", "/C", "prettier --stdin-filepath test.html"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := getCommand(tt.goos, tt.shell, tt.command)
			if cmd.Path != tt.wantPath {
				t.Errorf("got path %q, want %q", cmd.Path, tt.wantPath)
			}
			if len(cmd.Args) != len(tt.wantArgs) {
				t.Errorf("got %d args, want %d", len(cmd.Args), len(tt.wantArgs))
			}
			for i, arg := range cmd.Args {
				if i < len(tt.wantArgs) && arg != tt.wantArgs[i] {
					t.Errorf("arg %d: got %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}
