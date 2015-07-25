// Package parser defines an interface for parsers which the base template conforms to
package parser

type FuncMap map[string]interface{}

// A parser loads template files, and returns a template suitable for rendering content
type Parser interface {
	// Called before parsing commences for viewsPath
	StartParse(viewsPath string, helpers FuncMap) error

	// Can this parser handle this file?
	CanParseFile(path string) bool

	// Parse the file given and return a compiled template
	NewTemplate(path string) (Template, error)

	// Called when parsing finishes for viewsPath
	FinishParse(viewsPath string) error
}
