package parser

import (
	"fmt"
	"io"
	"regexp"
	got "text/template"
)

var textTemplateSet *got.Template
var textTemplateInclude *regexp.Regexp

// A HTML template using go HTML/template - NB this is not escaped and unsafe in HTML.
type TextTemplate struct {
	BaseTemplate
}

// Perform setup before parsing templates
func (t *TextTemplate) StartParse(viewsPath string, helpers FuncMap) error {
	textTemplateSet = got.New("").Funcs(got.FuncMap(helpers))

	// e.g. {{ template "shared/header.html" . }}
	textTemplateInclude = regexp.MustCompile(`{{\s*template\s*["]([\S]*)["].*}}`)
	return nil
}

// Can this parser handle this file path?
// test.csv.gotext
func (t *TextTemplate) CanParseFile(path string) bool {
	allowed := []string{".text.got",".csv.got"}
	return suffixes(path, allowed)
}

func (t *TextTemplate) NewTemplate(path string) (Template, error) {
	template := new(TextTemplate)
	err := template.Parse(path)
	return template, err
}

// Parse the template
func (t *TextTemplate) Parse(path string) error {
	err := t.BaseTemplate.Parse(path)

	// Add to our template set
	if textTemplateSet.Lookup(t.path) == nil {
		_, err = textTemplateSet.New(t.path).Parse(t.Source())
	} else {
		err = fmt.Errorf("Duplicate template:%s %s", t.Path(), t.Source())
	}

	return err
}

// Parse a string template
func (t *TextTemplate) ParseString(s string) error {
	err := t.BaseTemplate.ParseString(s)

	// Add to our template set
	if textTemplateSet.Lookup(t.Path()) == nil {
		_, err = textTemplateSet.New(t.path).Parse(t.Source())
	} else {
		err = fmt.Errorf("Duplicate template:%s %s", t.Path(), t.Source())
	}

	return err
}

// Finalise the template set, called after parsing is complete
// Record a list of dependent templates (for breaking caches automatically)
func (t *TextTemplate) Finalize(templates map[string]Template) error {

	// Search source for {{\s template "|`xxx`|" x }} pattern
	paths := textTemplateInclude.FindAllStringSubmatch(t.Source(), -1)

	// For all includes found, add the template to our dependency list
	for _, p := range paths {
		d := templates[p[1]]
		if d != nil {
			t.dependencies = append(t.dependencies, d)
		}
	}

	return nil
}

// BaseTemplate renders the template ignoring conHTML
func (t *TextTemplate) Render(writer io.Writer, context map[string]interface{}) error {
	tmpl := t.goTemplate()
	if tmpl == nil {
		return fmt.Errorf("Error rendering template:%s %s", t.Path(), t.Source())
	}

	return tmpl.Execute(writer, context)
}

func (t *TextTemplate) goTemplate() *got.Template {
	return textTemplateSet.Lookup(t.Path())
}
