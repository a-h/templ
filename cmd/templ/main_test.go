package main

import (
	"bytes"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expected     string
		expectedCode int
	}{
		{
			name:         "no args prints usage",
			args:         []string{},
			expected:     usageText,
			expectedCode: 0,
		},
		{
			name:         `"templ help" prints help`,
			args:         []string{"templ", "help"},
			expected:     usageText,
			expectedCode: 0,
		},
		{
			name:         `"templ --help" prints help`,
			args:         []string{"templ", "--help"},
			expected:     usageText,
			expectedCode: 0,
		},
		{
			name:         `"templ version" prints version`,
			args:         []string{"templ", "version"},
			expected:     templ.Version() + "\n",
			expectedCode: 0,
		},
		{
			name:         `"templ --version" prints version`,
			args:         []string{"templ", "--version"},
			expected:     templ.Version() + "\n",
			expectedCode: 0,
		},
		{
			name:         `"templ migrate" prints usage`,
			args:         []string{"templ", "migrate"},
			expected:     migrateUsageText,
			expectedCode: 0,
		},
		{
			name:         `"templ migrate --help" prints usage`,
			args:         []string{"templ", "migrate", "--help"},
			expected:     migrateUsageText,
			expectedCode: 0,
		},
		{
			name:         `"templ fmt --help" prints usage`,
			args:         []string{"templ", "fmt", "--help"},
			expected:     fmtUsageText,
			expectedCode: 0,
		},
		{
			name:         `"templ generate --help" prints usage`,
			args:         []string{"templ", "generate", "--help"},
			expected:     generateUsageText,
			expectedCode: 0,
		},
		{
			name:         `"templ lsp --help" prints usage`,
			args:         []string{"templ", "lsp", "--help"},
			expected:     lspUsageText,
			expectedCode: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := bytes.NewBuffer(nil)
			actualCode := run(actual, test.args)

			if actualCode != test.expectedCode {
				t.Errorf("expected code %v, got %v", test.expectedCode, actualCode)
			}
			if diff := cmp.Diff(test.expected, actual.String()); diff != "" {
				t.Error(diff)
				t.Error("expected:")
				t.Error(test.expected)
				t.Error("actual:")
				t.Error(actual.String())
			}
		})
	}
}
