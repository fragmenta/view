// Package view provides methods for rendering templates, and helper functions for golang views
package view

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fragmenta/view/helpers"
	"github.com/fragmenta/view/parser"
)

// The template sets, loaded on startup for production and on every request for development
var Templates map[string]parser.Template

// Helper functions available in templates
var Helpers map[string]interface{}

var Production bool

func DefaultHelpers() parser.FuncMap {
	funcs := make(parser.FuncMap)

	// HEAD helpers
	funcs["style"] = helpers.Style
	funcs["script"] = helpers.Script
	funcs["dev"] = func() bool { return !Production }

	// HTML helpers
	funcs["html"] = helpers.Html
	funcs["htmlattr"] = helpers.HtmlAttribute
	funcs["url"] = helpers.Url

	funcs["sanitize"] = helpers.Sanitize
	funcs["strip"] = helpers.Strip
	funcs["truncate"] = helpers.Truncate

	// Form helpers
	funcs["field"] = helpers.Field
	funcs["datefield"] = helpers.DateField
	funcs["textarea"] = helpers.TextArea
	funcs["select"] = helpers.Select
	funcs["selectarray"] = helpers.SelectArray
	funcs["optionsforselect"] = helpers.OptionsForSelect

	funcs["utcdate"] = helpers.UTCDate
	funcs["utcnow"] = helpers.UTCNow
	funcs["date"] = helpers.Date
	funcs["time"] = helpers.Time
	funcs["numberoptions"] = helpers.NumberOptions

	// CSV helper
	funcs["csv"] = helpers.CSV

	// String helpers
	funcs["blank"] = helpers.Blank

	// Math helpers
	funcs["mod"] = helpers.Mod
	funcs["odd"] = helpers.Odd
	funcs["add"] = helpers.Add

	// Array functions
	funcs["array"] = helpers.Array
	funcs["append"] = helpers.Append
	funcs["contains"] = helpers.Contains

	// Map functions
	funcs["map"] = helpers.Map
	funcs["set"] = helpers.Set
	funcs["setif"] = helpers.SetIf
	funcs["empty"] = helpers.Empty

	// Numeric helpers
	funcs["centstoprice"] = helpers.CentsToPrice
	funcs["pricetocents"] = helpers.PriceToCents

	// FIXME - deprecated, remove these
	funcs["counter"] = helpers.Counter
	funcs["counter_reset"] = helpers.CounterReset
	funcs["link_to_attr"] = helpers.LinkToAttributes
	funcs["link_to"] = helpers.LinkTo

	return funcs
}

func LoadTemplates() error {

	// Set up our helper functions if necessary
	if Helpers == nil {
		Helpers = DefaultHelpers()
	}

	// Scan all templates within the directories under us
	scanner, err := parser.NewScanner()
	if err != nil {
		return err
	}

	err = scanner.ScanPath("./src", Helpers)
	if err != nil {
		return err
	}

	Templates = scanner.Templates

	return nil
}

// Print out our list of templates for debug
func PrintTemplates() {
	for k, _ := range Templates {
		fmt.Printf("Template %s", k)
	}
	fmt.Printf("Finished scan of templates")
}

// Render a status code for the user, using a default template (if available)
func RenderStatus(writer http.ResponseWriter, status int) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)

	var title, message string

	switch status {

	case http.StatusUnauthorized, http.StatusForbidden:
		title = "Unauthorized"
		message = "Sorry, you don't have permission to perform that action."

	case http.StatusInternalServerError:
		title = "Server Error"
		message = "Sorry, an error occurred. Please let us know."

	case http.StatusNotFound:
		title = "Not Found"
		message = "Sorry, we couldn't find the requested page. If you think this was an error, please let us know."

	case http.StatusTeapot:
		title = "Teapot!"
		message = "I'm a little teapot, short and stout."

	default:
		title = "Oops"
		message = "Sorry, something went wrong, please let us know"
	}

	// Template name
	statusTemplate := fmt.Sprintf("app/views/%d.html.got", status)
	t := Templates[statusTemplate]
	if t != nil {
		c := map[string]interface{}{
			"title":   title,
			"message": message,
			"status":  status,
		}
		err := t.Render(writer, c)
		if err == nil {
			return
		}
	}

	// If not or error render a simple error page
	io.WriteString(writer, fmt.Sprintf("<h1>%s</h1><p>%s</p><p>Status:%d</p>", title, message, status))

}
