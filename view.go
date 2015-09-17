// Package view provides methods for rendering templates, and helper functions for golang views
package view

import (
	"fmt"
	"sync"

	"github.com/fragmenta/view/helpers"
	"github.com/fragmenta/view/parser"
)

// The template sets, loaded on startup for production and on every request for development
var Templates map[string]parser.Template

// This mutex guards the above Templates variable during reload
var mu sync.Mutex

// Helper functions available in templates
var Helpers map[string]interface{}

// Production is true if this server is running in production mode
var Production bool

func init() {
	// Set up our helper functions
	Helpers = DefaultHelpers()
}

// DefaultHelpers returns a default set of helpers for the app, which can then be extended/replaced
// NB if you change helper functions the templates must be reloaded at least once afterwards
func DefaultHelpers() parser.FuncMap {
	funcs := make(parser.FuncMap)

	// HEAD helpers
	funcs["style"] = helpers.Style
	funcs["script"] = helpers.Script
	funcs["dev"] = func() bool { return !Production }

	// HTML helpers
	funcs["html"] = helpers.HTML
	funcs["htmlattr"] = helpers.HTMLAttribute
	funcs["url"] = helpers.URL

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
	funcs["utctime"] = helpers.UTCTime
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

	// Numeric helpers - clean up and accept currency and other options in centstoprice
	funcs["centstobase"] = helpers.CentsToBase
	funcs["centstoprice"] = helpers.CentsToPrice
	funcs["pricetocents"] = helpers.PriceToCents

	return funcs
}

// LoadTemplates loads our templates, and assigns them to the package variable Templates
func LoadTemplates() error {

	// Scan all templates within the directories under us
	scanner, err := parser.NewScanner()
	if err != nil {
		return err
	}

	err = scanner.ScanPath("./src", Helpers)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	Templates = scanner.Templates

	return nil
}

// PrintTemplates prints out our list of templates for debug
func PrintTemplates() {
	for k := range Templates {
		fmt.Printf("Template %s", k)
	}
	fmt.Printf("Finished scan of templates")
}
