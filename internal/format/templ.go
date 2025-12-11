package format

import (
	"bytes"
	"fmt"

	"github.com/a-h/templ/internal/imports"
	"github.com/a-h/templ/internal/prettier"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
)

type Config struct {
	// PrettierCommand is the command to run to format the content of script and style elements.
	// If empty, the default is "prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME".
	PrettierCommand string
	// PrettierRequired indicates that formatting using Prettier must be applied.
	PrettierRequired bool
	// NormalizeSpacing enables consistent blank line spacing around component blocks.
	NormalizeSpacing bool
	// NormalizeOperators enables standardized spacing around + operators.
	NormalizeOperators bool
	// NormalizeHTMLTags enables conversion of XML-style self-closing tags to HTML5 style.
	NormalizeHTMLTags bool
	// CleanupBlankLines enables removal of excessive blank lines.
	CleanupBlankLines bool
}

// Templ formats templ source, returning the formatted output, whether it changed, and an error if any.
// The fileName is used for Go import processing, use an empty name if the source is not from a file.
func Templ(src []byte, fileName string, config Config) (output []byte, changed bool, err error) {
	t, err := parser.ParseString(string(src))
	if err != nil {
		return nil, false, err
	}
	t.Filepath = fileName
	t, err = imports.Process(t)
	if err != nil {
		return nil, false, err
	}

	if err = applyPrettier(t, config); err != nil {
		return nil, false, err
	}

	w := new(bytes.Buffer)
	if err = t.Write(w); err != nil {
		return nil, false, fmt.Errorf("formatting error: %w", err)
	}
	out := w.Bytes()

	// Apply post-parse formatting passes
	out = applyFormattingPasses(out, config)

	changed = !bytes.Equal(src, out)
	return out, changed, nil
}

// applyFormattingPasses applies formatting passes to normalize spacing, operators, and HTML tags.
func applyFormattingPasses(src []byte, config Config) []byte {
	result := src

	// Pass 1: Normalize spacing around component blocks
	if config.NormalizeSpacing {
		result = normalizeComponentSpacing(result)
	}

	// Pass 2: Normalize operator spacing
	if config.NormalizeOperators {
		result = normalizeOperatorSpacing(result)
	}

	// Pass 3: Normalize HTML tags
	if config.NormalizeHTMLTags {
		result = normalizeHTMLTags(result)
	}

	// Pass 4: Clean up blank lines
	if config.CleanupBlankLines {
		result = cleanupBlankLines(result)
	}

	return result
}

func applyPrettier(t *parser.TemplateFile, config Config) (err error) {
	// Check to see if prettier can be run.
	if config.PrettierCommand == "" {
		config.PrettierCommand = prettier.DefaultCommand
	}
	if !prettier.IsAvailable(config.PrettierCommand) {
		if config.PrettierRequired {
			return fmt.Errorf("prettier command %q is not available, please install it or set a different command using the -prettier-command flag", config.PrettierCommand)
		}
		// Prettier is not available, skip applying it.
		return nil
	}

	nodeFormatter := visitor.New()
	// Calculate the depth of each ScriptElement and RawElement in the tree so that the formatting is properly indented.
	nodeToDepth := make(map[parser.Node]int)
	nodeFormatter.HTMLTemplate = func(n *parser.HTMLTemplate) error {
		// Visit the children first to calculate their depth.
		for _, child := range n.Children {
			calculateNodeDepth(child, nodeToDepth, 1)
		}
		// Now that we have the depth of each node, we can format them.
		for _, child := range n.Children {
			if err := child.Visit(nodeFormatter); err != nil {
				return err
			}
		}
		return nil
	}
	nodeFormatter.ScriptElement = func(se *parser.ScriptElement) error {
		depth := nodeToDepth[se]
		return ScriptElement(se, depth, config.PrettierCommand)
	}
	nodeFormatter.RawElement = func(re *parser.RawElement) error {
		depth := nodeToDepth[re]
		if re.Name != "style" {
			return nil
		}
		return StyleElement(re, depth, config.PrettierCommand)
	}

	return nodeFormatter.VisitTemplateFile(t)
}

func calculateNodeDepth(e parser.Node, nodeToDepth map[parser.Node]int, depth int) {
	switch e := e.(type) {
	case *parser.ScriptElement:
		nodeToDepth[e] = depth
	case *parser.RawElement:
		nodeToDepth[e] = depth
	case parser.CompositeNode:
		for _, child := range e.ChildNodes() {
			calculateNodeDepth(child, nodeToDepth, depth+1)
		}
	}
}
