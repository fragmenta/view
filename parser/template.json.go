package parser

import (
	"fmt"
	got "html/template"
	"io"
	"regexp"
)

var jsonTemplateSet *got.Template
var jsonTemplateInclude *regexp.Regexp

// A JSON template using go HTML/template - NB this is not escaped and unsafe in HTML.
type JSONTemplate struct {
	BaseTemplate
}

// Perform setup before parsing templates
func (t *JSONTemplate) StartParse(viewsPath string, helpers FuncMap) error {
	jsonTemplateSet = got.New("").Funcs(got.FuncMap(helpers))
	// e.g. {{ template "shared/header.html" . }}
	jsonTemplateInclude = regexp.MustCompile(`{{\s*template\s*["]([\S]*)["].*}}`)
	return nil
}

// Can this parser handle this file path?
// test.csv.got
func (t *JSONTemplate) CanParseFile(path string) bool {
	allowed := []string{".json.got"}
	return suffixes(path, allowed)
}

func (t *JSONTemplate) NewTemplate(path string) (Template, error) {
	template := new(JSONTemplate)
	err := template.Parse(path)
	return template, err
}

// Parse the template
func (t *JSONTemplate) Parse(path string) error {
	err := t.BaseTemplate.Parse(path)

	// Add to our template set
	if jsonTemplateSet.Lookup(t.Path()) == nil {
		_, err = jsonTemplateSet.New(t.path).Parse(t.Source())
	} else {
		err = fmt.Errorf("Duplicate template:%s %s", t.Path(), t.Source())
	}

	return err
}

// Parse a string template
func (t *JSONTemplate) ParseString(s string) error {
	err := t.BaseTemplate.ParseString(s)

	// Add to our template set
	if jsonTemplateSet.Lookup(t.Path()) == nil {
		_, err = jsonTemplateSet.New(t.path).Parse(t.Source())
	} else {
		err = fmt.Errorf("Duplicate template:%s %s", t.Path(), t.Source())
	}

	return err
}

// Finalise the template set, called after parsing is complete
func (t *JSONTemplate) Finalize(templates map[string]Template) error {

	// Go html/template records dependencies both ways (child <-> parent)
	// tmpl.Templates() includes tmpl and children and parents
	// we only want includes listed as dependencies
	// so just do a simple search of parsed source instead

	// Search source for {{\s template "|`xxx`|" x }} pattern
	paths := jsonTemplateInclude.FindAllStringSubmatch(t.Source(), -1)

	// For all includes found, add the template to our dependency list
	for _, p := range paths {
		d := templates[p[1]]
		if d != nil {
			t.dependencies = append(t.dependencies, d)
		}
	}

	return nil
}

func (t *JSONTemplate) Render(writer io.Writer, context map[string]interface{}) error {
	return t.got().Execute(writer, context)
}

func (t *JSONTemplate) got() *got.Template {
	return jsonTemplateSet.Lookup(t.Path())
}
