// Package parser defines an interface for parsers (creating templates) and templates (rendering content), and defines a base template type which conforms to both interfaces and can be included in any templates
package parser

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Scanner scans paths for templates and creates a representation of each using parsers
type Scanner struct {
	// A map of all templates keyed by path name
	Templates map[string]Template

	// A set of parsers (in order) with which to parse templates
	Parsers []Parser

	// A set of paths (in order) from which to load templates
	Paths []string
}

func NewScanner() (*Scanner, error) {
	// Create a new template scanner
	s := &Scanner{}

	s.Templates = map[string]Template{}
	s.Parsers = []Parser{
        new(JSONTemplate),
		new(HTMLTemplate),
		new(TextTemplate),
	}

	return s, nil
}

// Scan a path for template files, including sub-paths
func (s *Scanner) ScanPath(root string, helpers FuncMap) error {

	s.Paths = append(s.Paths, root)

	root = path.Clean(root)

	// Store current path, and change to root path
	// so that template includes use relative paths from root
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(root)
	if err != nil {
		return err
	}

	// Set up parsers for this path
	for _, p := range s.Parsers {
		p.StartParse(root, helpers)
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		// Deal with files, directories we return nil error to recurse on them
		if !info.IsDir() {
			// Ask parsers in turn to handle the file - first one to claim it wins
			for _, p := range s.Parsers {
				if p.CanParseFile(path) {

					t, err := p.NewTemplate(path)
					if err != nil {
						return err
					}

					s.Templates[path] = t
					return nil
				}
			}

		}

		return nil
	})

	if err != nil {
		return err
	}

	// Finalise templates (build dependency list etc)
	// Supply them with a full set of templates to do this with
	// Alternative is to adjust mustache implementation here to give it idea of template set
	for _, t := range s.Templates {
		t.Finalize(s.Templates)
	}

	// Finalise parsers
	for _, p := range s.Parsers {
		p.FinishParse(root)
	}

	// Change back to original path
	err = os.Chdir(pwd)
	if err != nil {
		return err
	}

	return nil
}

// Rescan all template paths
func (s *Scanner) RescanPaths(helpers FuncMap) error {
	// Make sure templates is empty
	s.Templates = make(map[string]Template)

	// Scan paths again
	for _, p := range s.Paths {
		err := s.ScanPath(p, helpers)
		if err != nil {
			return err
		}
	}

	return nil
}

// PATH UTILITIES

// Is the file path supplied a dot file?
func dotFile(p string) bool {
	return strings.HasPrefix(path.Base(p), ".")
}

// Does the path have this suffix (ignoring dotfiles)?
func suffix(p string, suffix string) bool {
	if dotFile(p) {
		return false
	}
	return strings.HasSuffix(p, suffix)
}

// Does the path have these suffixes (ignoring dotfiles)?
func suffixes(p string, suffixes []string) bool {
	if dotFile(p) {
		return false
	}

	for _, s := range suffixes {
		if strings.HasSuffix(p, s) {
			return true
		}
	}

	return false
}
