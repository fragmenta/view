package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// Renderer is a view which is set up on each request and renders the response to its writer
type Renderer struct {

	// The view rendering context
	context map[string]interface{}

	// The writer to write the context to
	writer http.ResponseWriter

	// The layout template to render in
	layout string

	// The template to render
	template string

	// The format to render with (html,json etc)
	format string

	// The status to render with
	status int

	// The request path
	path string
}

// RenderContext is the type passed in to New, which helps construct the rendering view
// Alternatively, you can use NewWithPath, which doesn't require a RenderContext
type RenderContext interface {
	Path() string
	RenderContext() map[string]interface{}
	Writer() http.ResponseWriter
}

// New creates a new Renderer
func New(c RenderContext) *Renderer {
	r := &Renderer{
		path:     c.Path(),
		layout:   "app/views/layout.html.got",
		template: "",
		format:   "text/html",
		status:   http.StatusOK,
		context:  c.RenderContext(),
		writer:   c.Writer(),
	}

	// This sets layout and template based on the view.path
	r.setDefaultTemplates()

	return r
}

// NewWithPath creates a new Renderer with a path and an http.ResponseWriter
func NewWithPath(p string, w http.ResponseWriter) *Renderer {
	r := &Renderer{
		path:     p,
		layout:   "app/views/layout.html.got",
		template: "",
		format:   "text/html",
		status:   http.StatusOK,
		context:  make(map[string]interface{}, 0),
		writer:   w,
	}

	// This sets layout and template based on the view.path
	r.setDefaultTemplates()

	return r
}

// Layout sets the layout used
func (r *Renderer) Layout(layout string) *Renderer {
	r.layout = layout
	return r
}

// Template sets the template used
func (r *Renderer) Template(template string) *Renderer {
	r.template = template
	return r
}

// Format sets the format used, e.g. text/html,
func (r *Renderer) Format(format string) *Renderer {
	r.format = format
	return r
}

// Path sets the request path on the renderer (used for choosing a default template)
func (r *Renderer) Path(p string) *Renderer {
	r.path = path.Clean(p)
	return r
}

// Status returns the Renderer status
func (r *Renderer) Status(status int) *Renderer {
	r.status = status
	return r
}

// Text sets the view content as text
func (r *Renderer) Text(content string) *Renderer {
	r.context["content"] = content
	return r
}

// HTML sets the view content as html (use with caution)
func (r *Renderer) HTML(content string) *Renderer {
	r.context["content"] = template.HTML(content)
	return r
}

// AddKey adds a key/value pair to context
func (r *Renderer) AddKey(key string, value interface{}) *Renderer {
	r.context[key] = value
	return r
}

// Context sets the entire context for rendering
func (r *Renderer) Context(c map[string]interface{}) *Renderer {
	r.context = c
	return r
}

// RenderToString renders our template into layout using our context and return a string
func (r *Renderer) RenderToString() (string, error) {

	content := ""

	if len(r.template) > 0 {
		mu.RLock()
		t := scanner.Templates[r.template]
		mu.RUnlock()
		if t == nil {
			return content, fmt.Errorf("No such template found %s", r.template)
		}

		var rendered bytes.Buffer
		err := t.Render(&rendered, r.context)
		if err != nil {
			return content, err
		}

		content = rendered.String()
	}

	return content, nil
}

// Render our template into layout using our context and write out to writer
func (r *Renderer) Render() error {

	// Reload if not in production
	if !Production {
		fmt.Printf("#warn Reloading templates in development mode\n")
		err := ReloadTemplates()
		if err != nil {
			return err
		}
	}

	// If we have a template, render it
	// using r.Context unless overridden by content being set with .Text("My string")
	if len(r.template) > 0 && r.context["content"] == nil {
		mu.RLock()
		t := scanner.Templates[r.template]
		mu.RUnlock()
		if t == nil {
			return fmt.Errorf("#error No such template found %s", r.template)
		}

		var rendered bytes.Buffer
		err := t.Render(&rendered, r.context)
		if err != nil {
			return fmt.Errorf("#error Could not render template %s - %s", r.template, err)
		}

		if r.layout != "" {
			r.context["content"] = template.HTML(rendered.String())
		} else {
			r.context["content"] = rendered.String()
		}
	}

	// Now render the content into the layout template
	if r.layout != "" {
		mu.RLock()
		layout := scanner.Templates[r.layout]
		mu.RUnlock()
		if layout == nil {
			return fmt.Errorf("#error Could not find layout %s", r.layout)
		}

		err := layout.Render(r.writer, r.context)
		if err != nil {
			return fmt.Errorf("#error Could not render layout %s %s", r.layout, err)
		}

	} else if r.context["content"] != nil {
		// Deal with no layout by rendering content directly to writer
		_, err := io.WriteString(r.writer, r.context["content"].(string))
		r.writer.Header().Set("Content-Type", r.format+"; charset=utf-8")
		r.writer.WriteHeader(r.status)
		return err
	}

	return nil
}

// Set sensible default layout/template paths after we know our path
// /pages => pages/views/index.html.got
// /pages/create => pages/views/create.html.got
// /pages/xxx => pages/views/show.html.got
// /pages/xxx/edit => pages/views/edit.html.got
func (r *Renderer) setDefaultTemplates() {

	// First deal with home (a special case)
	if r.path == "/" {
		r.template = "pages/views/home.html.got"
		return
	}

	// Now see if we can find a template based on our path
	trimmed := strings.Trim(r.path, "/")
	parts := strings.Split(trimmed, "/")

	pkg := "app"
	action := "index"

	// TODO: add handling for theme templates
	// we should attempt to match theme paths first, before default paths
	// but need to know which theme is active for the domain for each request

	// Deal with default paths by matching the path within the folders
	switch len(parts) {
	default:
	case 1: // /pages
		pkg = parts[0]
	case 2: // /pages/create or /pages/1 etc
		pkg = parts[0]
		action = parts[1]
		// NB the +, we require 1 or more digits
		numeric, _ := regexp.MatchString("^[0-9]+", parts[1])
		if numeric {
			action = "show"
		}
	case 3: // /pages/xxx/edit
		pkg = parts[0]
		action = parts[2]
	}

	//	fmt.Printf("#templates setting default template:%s/views/%s.html.got", pkg, action)

	// Set a default template
	mu.RLock()
	path := fmt.Sprintf("%s/views/%s.html.got", pkg, action)
	if scanner.Templates[path] != nil {
		r.template = path
	}

	// Set a default layout
	path = fmt.Sprintf("%s/views/layout.html.got", pkg)
	if scanner.Templates[path] != nil {
		r.layout = path
	}
	mu.RUnlock()
}
