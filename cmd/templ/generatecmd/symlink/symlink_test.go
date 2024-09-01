package symlink

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/cmd/templ/testproject"
)

func TestSymlink(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	t.Run("can generate if root is symlink", func(t *testing.T) {
		// templ generate -f templates.templ
		dir, err := testproject.Create("github.com/a-h/templ/cmd/templ/testproject")
		if err != nil {
			t.Fatalf("failed to create test project: %v", err)
		}
		defer os.RemoveAll(dir)

		symlinkPath := dir + "-symlink"
		err = os.Symlink(dir, symlinkPath)
		if err != nil {
			t.Fatalf("failed to create dir symlink: %v", err)
		}
		defer os.Remove(symlinkPath)

		// Delete the templates_templ.go file to ensure it is generated.
		err = os.Remove(path.Join(symlinkPath, "templates_templ.go"))
		if err != nil {
			t.Fatalf("failed to remove templates_templ.go: %v", err)
		}

		// Run the generate command.
		err = generatecmd.Run(context.Background(), log, generatecmd.Arguments{
			Path: symlinkPath,
		})
		if err != nil {
			t.Fatalf("failed to run generate command: %v", err)
		}

		// Check the templates_templ.go file was created.
		_, err = os.Stat(path.Join(symlinkPath, "templates_templ.go"))
		if err != nil {
			t.Fatalf("templates_templ.go was not created: %v", err)
		}
	})
}
