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

	// Helpers is a list of helper functions
	Helpers FuncMap
}

// NewScanner creates a new template scanner
func NewScanner(paths []string, helpers FuncMap) (*Scanner, error) {
	s := &Scanner{
		Helpers:   helpers,
		Paths:     paths,
		Templates: make(map[string]Template),
		Parsers:   []Parser{new(JSONTemplate), new(HTMLTemplate), new(TextTemplate)},
	}

	return s, nil
}

// ScanPath scans a path for template files, including sub-paths
func (s *Scanner) ScanPath(root string) error {

	s.Paths = append(s.Paths, root)

	root = path.Clean(root)

	// Store current path, and change to root path
	// so that template includes use relative paths from root
	// this may not be necc. any more, test removing it
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(root)
	if err != nil {
		return err
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

					fullpath := filepath.Join(root, path)
					t, err := p.NewTemplate(fullpath, path)
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

	// Change back to original path
	err = os.Chdir(pwd)
	if err != nil {
		return err
	}

	return nil
}

// ScanPaths resets template list and rescans all template paths
func (s *Scanner) ScanPaths() error {
	// Make sure templates is empty
	s.Templates = make(map[string]Template)

	// Set up the parsers
	for _, p := range s.Parsers {
		err := p.Setup(s.Helpers)
		if err != nil {
			return err
		}
	}

	// Scan paths again
	for _, p := range s.Paths {
		err := s.ScanPath(p)
		if err != nil {
			return err
		}
	}

	// Now parse and finalize templates
	for _, t := range s.Templates {
		err := t.Parse()
		if err != nil {
			return err
		}
	}

	// Now finalize templates
	for _, t := range s.Templates {
		err := t.Finalize(s.Templates)
		if err != nil {
			return err
		}
	}

	return nil
}

// PATH UTILITIES

// dotFile returns true if the file path supplied a dot file?
func dotFile(p string) bool {
	return strings.HasPrefix(path.Base(p), ".")
}

// suffix returns true if the path have this suffix (ignoring dotfiles)?
func suffix(p string, suffix string) bool {
	if dotFile(p) {
		return false
	}
	return strings.HasSuffix(p, suffix)
}

// suffixes returns true if the path has these suffixes (ignoring dotfiles)?
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
