package prettier

import (
	"fmt"
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
			t.Errorf("IsAvailable should return true for existing command %q", DefaultCommand())
		}
	})
}

func TestElementReturnsContentOnly(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		depth           int
		prettierCommand string
		expected        string
	}{
		{
			name:            "unchanged content at depth 0 returns original content",
			content:         "\n\tconsole.log(\"hello\");\n",
			depth:           0,
			prettierCommand: "cat",
			expected:        "\n\tconsole.log(\"hello\");\n",
		},
		{
			name:            "unchanged content at depth 1 returns original content",
			content:         "\n\tconsole.log(\"hello\");\n",
			depth:           1,
			prettierCommand: "cat",
			expected:        "\n\tconsole.log(\"hello\");\n",
		},
		{
			name:            "unchanged content at depth 2 returns original content",
			content:         "\n\tconsole.log(\"hello\");\n",
			depth:           2,
			prettierCommand: "cat",
			expected:        "\n\tconsole.log(\"hello\");\n",
		},
	}

	prettierAvailable := IsAvailable(DefaultCommand())
	if prettierAvailable {
		tests = append(tests, []struct {
			name            string
			content         string
			depth           int
			prettierCommand string
			expected        string
		}{
			{
				name:            "simple console.log at depth 0",
				content:         "\n\tconsole.log(\"Hello, World!\");\n",
				depth:           0,
				prettierCommand: DefaultCommand(),
			},
			{
				name:            "simple console.log at depth 2",
				content:         "\n\t\tconsole.log(\"Hello, World!\");\n\t",
				depth:           2,
				prettierCommand: DefaultCommand(),
			},
			{
				name:            "functions with empty bodies at depth 1",
				content:         "\n\t\tfunction func1(e) {\n\t\t}\n\t\tfunction func2() {\n\t\t}\n\t",
				depth:           1,
				prettierCommand: DefaultCommand(),
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Element("script", "", tt.content, tt.depth, tt.prettierCommand)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expected != "" && got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
			if strings.Contains(got, "data-templ-depth") {
				t.Errorf("data-templ-depth leaked into output:\n%s", got)
			}
			if strings.Contains(got, "<script>") || strings.Contains(got, "</script>") {
				t.Errorf("script tags leaked into content:\n%s", got)
			}
			if strings.Contains(got, "<div") || strings.Contains(got, "</div>") {
				t.Errorf("div tags leaked into content:\n%s", got)
			}
		})
	}
}

func TestDefaultCommand(t *testing.T) {
	bothAvailable := func(name string) (string, error) {
		if name == "prettier" || name == "prettierd" {
			return "/usr/bin/" + name, nil
		}
		return "", fmt.Errorf("not found")
	}
	onlyPrettier := func(name string) (string, error) {
		if name == "prettier" {
			return "/usr/bin/prettier", nil
		}
		return "", fmt.Errorf("not found")
	}
	onlyPrettierd := func(name string) (string, error) {
		if name == "prettierd" {
			return "/usr/bin/prettierd", nil
		}
		return "", fmt.Errorf("not found")
	}
	neitherAvailable := func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	tests := []struct {
		name     string
		shell    string
		lookPath func(string) (string, error)
		want     string
	}{
		{
			name:     "prefers prettier when both are available",
			shell:    "/bin/bash",
			lookPath: bothAvailable,
			want:     "prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME",
		},
		{
			name:     "uses prettier when only prettier is available",
			shell:    "/bin/bash",
			lookPath: onlyPrettier,
			want:     "prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME",
		},
		{
			name:     "falls back to prettierd when prettier is not available",
			shell:    "/bin/bash",
			lookPath: onlyPrettierd,
			want:     "prettierd --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME",
		},
		{
			name:     "returns prettier command when neither is available",
			shell:    "/bin/bash",
			lookPath: neitherAvailable,
			want:     "prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME",
		},
		{
			name:     "nushell uses prettierd when prettier is not available",
			shell:    "/usr/bin/nu",
			lookPath: onlyPrettierd,
			want:     "prettierd --use-tabs --stdin-filepath $env.TEMPL_PRETTIER_FILENAME",
		},
		{
			name:     "nushell prefers prettier when both are available",
			shell:    "/usr/bin/nu",
			lookPath: bothAvailable,
			want:     "prettier --use-tabs --stdin-filepath $env.TEMPL_PRETTIER_FILENAME",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultCommand(tt.shell, tt.lookPath)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
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
			command:  nuCommands["prettier"],
			wantPath: "/usr/bin/nu",
			wantArgs: []string{"/usr/bin/nu", "-c", nuCommands["prettier"]},
		},
		{
			name:     "bash uses the default posix command",
			goos:     "linux",
			shell:    "/bin/bash",
			command:  posixCommands["prettier"],
			wantPath: "/bin/bash",
			wantArgs: []string{"/bin/bash", "-c", posixCommands["prettier"]},
		},
		{
			name:     "zsh uses the default posix command",
			goos:     "linux",
			shell:    "/bin/zsh",
			command:  posixCommands["prettier"],
			wantPath: "/bin/zsh",
			wantArgs: []string{"/bin/zsh", "-c", posixCommands["prettier"]},
		},
		{
			name:     "empty shell defaults to the default posix command",
			goos:     "linux",
			shell:    "",
			command:  posixCommands["prettier"],
			wantPath: "/bin/sh",
			wantArgs: []string{"/bin/sh", "-c", posixCommands["prettier"]},
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
