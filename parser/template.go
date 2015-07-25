package parser

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
)

// A template renders its content given a ViewContext
type Template interface {
	// Parse a template file
	Parse(path string) error

	// Called after parsing is finished
	Finalize(templates map[string]Template) error

	// Render to this writer
	Render(writer io.Writer, context map[string]interface{}) error

	// Return the original template content
	Source() string

	// Return the template path
	Path() string

	// Return the cache key
	CacheKey() string

	// Return dependencies of this template (used for creating cache keys)
	Dependencies() []Template
}

var MaxCacheKeyLength = 250

// A base template which conforms to Template and Parser interfaces.
// This is an abstract base type, we use html or text templates
type BaseTemplate struct {
	path         string
	source       string     // at present we store in memory
	key          string     // set at parse time
	dependencies []Template // set at parse time
}

// PARSER

// Start parsing
func (t *BaseTemplate) StartParse(viewsPath string, helpers FuncMap) error {
	return nil
}

// Can parse file
func (t *BaseTemplate) CanParseFile(path string) bool {
	if dotFile(path) {
		return false
	}

	return true
}

// Return a newly created template for this path
func (t *BaseTemplate) NewTemplate(path string) (Template, error) {
	template := new(BaseTemplate)
	err := template.Parse(path)
	return template, err
}

// Finish parsing this path
func (t *BaseTemplate) FinishParse(viewsPath string) error {
	return nil
}

// TEMPLATE PARSING

// Parse the template (BaseTemplate simply stores it)
func (t *BaseTemplate) Parse(path string) error {
	t.path = path
	// Read the file
	s, err := t.readFile(path)
	if err == nil {
		t.source = s
	}

	return err
}

// Parse a string template
func (t *BaseTemplate) ParseString(s string) error {
	t.path = t.generateHash(s)
	t.source = s
	return nil
}

// BaseTemplate renders the template ignoring context
func (t *BaseTemplate) Render(writer io.Writer, context map[string]interface{}) error {
	writer.Write([]byte(t.Source()))
	return nil
}

// Called on each template after parsing is finished, supplying complete template set.
func (t *BaseTemplate) Finalize(templates map[string]Template) error {

	t.dependencies = []Template{}

	return nil
}

// Return the parsed version of this template
func (t *BaseTemplate) Source() string {
	return t.source
}

// Return the path of this template
func (t *BaseTemplate) Path() string {
	return t.path
}

// Return the cache key of this template -
// (this is generated from path + hash of contents + dependency hash keys).
// So it automatically changes when templates are changed
func (t *BaseTemplate) CacheKey() string {
	// If we have a key, return it
	// NB this relies on templates being reloaded on reload of app in production...
	if t.key != "" {
		return t.key
	}

	//println("Making key for",t.Path())

	// Otherwise generate the key
	t.key = t.path + "/" + t.generateHash(t.Source())

	for _, d := range t.dependencies {
		t.key = t.key + "-" + d.CacheKey()
	}

	// Finally, if the key is too long, set it to a hash of the key instead
	// (Memcache for example has limits on key length)
	// possibly we should deal with this at a higher level
	// I'd suggest always md5 keys with /view/ prefix...
	// put this into cache itself though...
	if len(t.key) > MaxCacheKeyLength {
		t.key = t.generateHash(t.key)
	}

	return t.key
}

// Return which other templates this one depends on (for generating nested cache keys)
func (t *BaseTemplate) Dependencies() []Template {
	return t.dependencies
}

// Utility method to read files into a string
func (t *BaseTemplate) readFile(path string) (string, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		println("Error reading template file at path ", path)
		return "", err
	}
	return string(fileBytes), err
}

// Utility method to generate a hash from string
func (t *BaseTemplate) generateHash(input string) string {
	h := md5.New()
	io.WriteString(h, input)
	return fmt.Sprintf("%x", h.Sum(nil))
}
