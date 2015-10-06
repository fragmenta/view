package parser

import (
	"fmt"
	"io"
	got "text/template"
)

var textTemplateSet *got.Template

// A HTML template using go HTML/template - NB this is not escaped and unsafe in HTML.
type TextTemplate struct {
	BaseTemplate
}

// Perform setup before parsing templates
func (t *TextTemplate) StartParse(viewsPath string, helpers FuncMap) error {
	textTemplateSet = got.New("").Funcs(got.FuncMap(helpers))
	return nil
}

// Can this parser handle this file path?
// test.csv.gotext
func (t *TextTemplate) CanParseFile(path string) bool {
	allowed := []string{".text.got", ".csv.got"}
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
	paths := templateInclude.FindAllStringSubmatch(t.Source(), -1)

	// For all includes found, add the template to our dependency list
	for _, p := range paths {
		d := templates[p[1]]
		if d != nil {
			t.dependencies = append(t.dependencies, d)
		}
	}

	return nil
}

// Render renders the template
func (t *TextTemplate) Render(writer io.Writer, context map[string]interface{}) error {
	tmpl := t.goTemplate()
	if tmpl == nil {
		return fmt.Errorf("Error rendering template:%s %s", t.Path(), t.Source())
	}

	return tmpl.Execute(writer, context)
}

// goTemplate returns teh underlying go template
func (t *TextTemplate) goTemplate() *got.Template {
	return textTemplateSet.Lookup(t.Path())
}
