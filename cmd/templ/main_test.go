package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
		expectedCode   int
	}{
		{
			name:           "no args prints usage",
			args:           []string{},
			expectedStderr: usageText,
			expectedCode:   64, // EX_USAGE
		},
		{
			name:           `"templ help" prints help`,
			args:           []string{"templ", "help"},
			expectedStdout: usageText,
			expectedCode:   0,
		},
		{
			name:           `"templ --help" prints help`,
			args:           []string{"templ", "--help"},
			expectedStdout: usageText,
			expectedCode:   0,
		},
		{
			name:           `"templ version" prints version`,
			args:           []string{"templ", "version"},
			expectedStdout: templ.Version() + "\n",
			expectedCode:   0,
		},
		{
			name:           `"templ --version" prints version`,
			args:           []string{"templ", "--version"},
			expectedStdout: templ.Version() + "\n",
			expectedCode:   0,
		},
		{
			name:           `"templ fmt --help" prints usage`,
			args:           []string{"templ", "fmt", "--help"},
			expectedStdout: fmtUsageText,
			expectedCode:   0,
		},
		{
			name:           `"templ lsp --help" prints usage`,
			args:           []string{"templ", "lsp", "--help"},
			expectedStdout: lspUsageText,
			expectedCode:   0,
		},
		{
			name:           `"templ info --help" prints usage`,
			args:           []string{"templ", "info", "--help"},
			expectedStdout: infoUsageText,
			expectedCode:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdin := strings.NewReader("")
			stdout := bytes.NewBuffer(nil)
			stderr := bytes.NewBuffer(nil)
			actualCode := run(stdin, stdout, stderr, test.args)

			if actualCode != test.expectedCode {
				t.Errorf("expected code %v, got %v", test.expectedCode, actualCode)
			}
			if diff := cmp.Diff(test.expectedStdout, stdout.String()); diff != "" {
				t.Error(diff)
				t.Error("expected stdout:")
				t.Error(test.expectedStdout)
				t.Error("actual stdout:")
				t.Error(stdout.String())
			}
			if diff := cmp.Diff(test.expectedStderr, stderr.String()); diff != "" {
				t.Error(diff)
				t.Error("expected stderr:")
				t.Error(test.expectedStderr)
				t.Error("actual stderr:")
				t.Error(stderr.String())
			}
		})
	}
}
