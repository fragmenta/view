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

    "github.com/fragmenta/view/helpers"
)

// The renderer should perhaps also have a log reference from router

// A new renderer is set up fresh for each request
type Renderer struct {

	// The view rendering context
	context map[string]interface{}

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

// The required context interface for setting up a view
type SetupContext interface {
    CurrentPath() string
    CurrentUser() interface {}
}

// A dummy context which supplies no info
type Empty struct {
}

func (m* Empty) CurrentPath() string  {
    return ""
}
func (m* Empty) CurrentUser() interface {}  {
    return nil
}


// Create a new view.Renderer
func New(c SetupContext) *Renderer {
	r := &Renderer{
		path:     c.CurrentPath(),
		layout:   "app/views/layout.html.got",
		template: "",
		format:   "text/html",
		status:   http.StatusOK,
		context:  make(map[string]interface{}),
	}
    
    r.context["current_user"] = c.CurrentUser()

	// This sets layout and template based on the view.path
	r.setDefaultTemplates()

	return r
}

func (r *Renderer) Layout(layout string) *Renderer {
	r.layout = layout
	return r
}

func (r *Renderer) Template(template string) *Renderer {
	r.template = template
	return r
}

// Format, e.g. text/html,
func (r *Renderer) Format(format string) *Renderer {
	r.format = format
	return r
}

// Set the path
func (r *Renderer) Path(p string) *Renderer {
	r.path = path.Clean(p)
	return r
}

// Status
func (r *Renderer) Status(status int) *Renderer {
	r.status = status
	return r
}

// Set the view content as text
func (r *Renderer) Text(content string) *Renderer {
	r.context["content"] = content
	return r
}

// Set the view content as html (use with caution)
func (r *Renderer) HTML(content string) *Renderer {
	r.context["content"] = template.HTML(content)
	return r
}

// Add a key/value pair to context
func (r *Renderer) AddKey(key string, value interface{}) *Renderer {
	r.context[key] = value
	return r
}


func (r *Renderer) Context(c map[string]interface{}) *Renderer {
	r.context = c
	return r
}


// Render our template into layout using our context and write out to writer
func (r *Renderer) RenderString() (string,error) {
    
    content := ""
  
    if len(r.template) > 0 {

		t := Templates[r.template]
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
func (r *Renderer) Render(writer http.ResponseWriter) error {

	// FIXME URGENT: What we should do here is of course only to reload those templates 
    // which have changed since we last reloaded, also need mutex on templates to avoid race condition
    // not a problem in dev but still worth fixing Causes 502 on concurrent loads 
    if !Production {
     //   fmt.Println("#info Reloading templates in development mode")
	//	LoadTemplates()
	}
   
   
	// If we have a template, render it
	// using r.Context unless overridden by content being set with .Text("My string")
	if len(r.template) > 0 && r.context["content"] == nil {

		t := Templates[r.template]
		if t == nil {
			err := fmt.Errorf("No such template found %s", r.template)
			r.RenderError(writer, err)
		
			return err
		}

		var rendered bytes.Buffer
		err := t.Render(&rendered, r.context)
		if err != nil {
			errfull := fmt.Errorf("Could not render template %s - %s", r.template, err)
			r.RenderError(writer, errfull)
		
			return err
		}
        
        if r.layout != "" {
            r.context["content"] = template.HTML(rendered.String())
        } else {
             r.context["content"] = rendered.String()
        }
	}

	// Now render the content into the layout template
	if r.layout != "" {
		layout := Templates[r.layout]

		// log.Printf("Using layout %v",Templates)

		if layout == nil {
			err := fmt.Errorf("Could not find layout %s in %s", r.layout, Templates)
			r.RenderError(writer, err)
			
			return err
		}

		err := layout.Render(writer, r.context)
		if err != nil {
			err := fmt.Errorf("Could not render layout %s %s", r.layout, err)
			r.RenderError(writer, err)
			
			return err
		}

	} else if r.context["content"] != nil {
		// Deal with no layout by rendering content directly to writer
		_, err := io.WriteString(writer, r.context["content"].(string) )
		writer.Header().Set("Content-Type", r.format+"; charset=utf-8")
		writer.WriteHeader(r.status)
        return err
	}

    // Reset our helpers on every render
    helpers.CounterReset()

	return nil
}

// Render our template into layout using our context and write out to writer
func (r *Renderer) RenderError(writer http.ResponseWriter, err error) {
    r.status = http.StatusInternalServerError
    
    // FIXME - need two things here - need debug status (to show errors + stack trace)
    // need log reference to log properly to log file
    
    // Need to log this here, but that's up to the app no?
    // If this were on router, we could log the error there, so consider moving it?
    fmt.Printf("#error %s\n",err)
    
	// Assign the error to the context so that the template can use it
    if !Production {
        r.context["error"] = err
    }
        
	// Check if app/views/500.html.got exists, use that with our view context
	t := Templates["app/views/500.html.got"]
	if t != nil {
    	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
    	writer.WriteHeader(r.status)

		err := t.Render(writer, r.context)
		if err != nil {
			fmt.Printf("ERROR on render error:%s\n",err)
            RenderStatus(writer,r.status)
		}
        
        return
	}

    // If not template fall back on render status for default error
	RenderStatus(writer,r.status)
}

func (r *Renderer) RenderStatus(writer http.ResponseWriter, status int) {
    r.status = status
    RenderStatus(writer,status)
}

/*

// There is an argument here for having our own special error type, with the following attributes:


Status int // required?
Error string // message (e.g. this column is too long)
Column string // for db errors


// The other way to deal is just to format messages into standard errors - I think I prefer that approach as it doesn't require yet another type 




// Render our template into layout using our context and write out to writer
func (r *Renderer) RenderErrors(writer http.ResponseWriter, status int, errors []error) {

	// For each error, add it to the context as a key?
	r.context["errors"] = errors

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)

	// Check if app/views/404.html.got exists, use that
	t := Templates["app/views/500.html.got"]
	if t != nil {
		err := t.Render(writer, r.context)
		if err == nil {
			return
		}
	}

	// If not or error render a simple error page with first error
	output := fmt.Sprintf("<h1>Error</h1><p>%s</p>", errors[0])
	io.WriteString(writer, output)

}
*/


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
		numeric, _ := regexp.MatchString("^[0-9]*$", parts[1])
		if numeric {
			action = "show"
		}
	case 3: // /pages/xxx/edit
		pkg = parts[0]
		action = parts[2]
	}

	// Set a default template
	path := fmt.Sprintf("%s/views/%s.html.got", pkg, action)
	if Templates[path] != nil {
		r.template = path
	}

	// Set a default layout
	path = fmt.Sprintf("%s/views/layout.html.got", pkg)
	if Templates[path] != nil {
		r.layout = path
	}

}
