package helpers

import (
	"bytes"
	"fmt"
	got "html/template"
	"io"
	"strings"

	parser "github.com/fragmenta/view/internal/html"
)

// NB all HTML with user input must be escaped, see https://www.owasp.org/index.php/XSS_Prevention_Cheatsheet

// These two should instead be using assets package?

// Style inserts a css tag
func Style(name string) got.HTML {
	return got.HTML(fmt.Sprintf("<link href=\"/assets/styles/%s.css\" media=\"all\" rel=\"stylesheet\" type=\"text/css\" />", EscapeURL(name)))
}

// Script inserts a script tag
func Script(name string) got.HTML {
	return got.HTML(fmt.Sprintf("<script src=\"/assets/scripts/%s.js\" type=\"text/javascript\"></script>", EscapeURL(name)))
}

// Escape escapes HTML using HTMLEscapeString
func Escape(s string) string {
	return got.HTMLEscapeString(s)
}

// EscapeURL escapes URLs using HTMLEscapeString
func EscapeURL(s string) string {
	return got.URLQueryEscaper(s)
}

// Link returns got.HTML with an anchor link given text and URL required
// Attributes (if supplied) should not contain user input
func Link(t string, u string, a ...string) got.HTML {
	attributes := ""
	if len(a) > 0 {
		attributes = strings.Join(a, " ")
	}
	return got.HTML(fmt.Sprintf("<a href=\"%s\" %s>%s</a>", Escape(u), Escape(attributes), Escape(t)))
}

// HTML returns a string (which must not contain user input) as go template HTML
func HTML(s string) got.HTML {
	return got.HTML(s)
}

// HTMLAttribute returns a string (which must not contain user input) as go template HTMLAttr
func HTMLAttribute(s string) got.HTMLAttr {
	return got.HTMLAttr(s)
}

// URL returns returns a string (which must not contain user input) as go template URL
func URL(s string) got.URL {
	return got.URL(s)
}

// Strip all html tags and returns as go template HTML
func Strip(s string) got.HTML {
	return Sanitize(s, []string{}, []string{})
}

// Sanitize sanitises html, allowing some tags using the html parser from golang.org/x/net/html and returns as go template HTML
// Usage: sanitize.HTMLAllowing("<b id=id>my html</b>",[]string{"b"},[]string{"id")
func Sanitize(s string, args ...[]string) got.HTML {

	var ignoreTags = []string{"title", "script", "style", "iframe", "frame", "frameset", "noframes", "noembed", "embed", "applet", "object"}
	var defaultTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "div", "span", "hr", "p", "br", "b", "i", "ol", "ul", "li", "strong", "em", "a", "img", "pre", "code"}
	var defaultAttributes = []string{"id", "class", "src", "title", "alt", "name", "rel", "href"}

	allowedTags := defaultTags
	if len(args) > 0 {
		allowedTags = args[0]
	}
	allowedAttributes := defaultAttributes
	if len(args) > 1 {
		allowedAttributes = args[1]
	}

	// Parse the html
	tokenizer := parser.NewTokenizer(strings.NewReader(s))

	buffer := bytes.NewBufferString("")
	ignore := ""

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		switch tokenType {

		case parser.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				return got.HTML(buffer.String())
			}

			fmt.Println("Error parsing html") // we should perhaps return an error
			return got.HTML("")

		case parser.StartTagToken:

			if len(ignore) == 0 && includes(allowedTags, token.Data) {
				token.Attr = cleanAttributes(token.Attr, allowedAttributes)
				buffer.WriteString(token.String())
			} else if includes(ignoreTags, token.Data) {
				ignore = token.Data
			}

		case parser.SelfClosingTagToken:

			if len(ignore) == 0 && includes(allowedTags, token.Data) {
				token.Attr = cleanAttributes(token.Attr, allowedAttributes)
				buffer.WriteString(token.String())
			} else if token.Data == ignore {
				ignore = ""
			}

		case parser.EndTagToken:
			if len(ignore) == 0 && includes(allowedTags, token.Data) {
				token.Attr = []parser.Attribute{}
				buffer.WriteString(token.String())
			} else if token.Data == ignore {
				ignore = ""
			}

		case parser.TextToken:
			// We allow text content through, unless ignoring this entire tag and its contents (including other tags)
			if ignore == "" {
				buffer.WriteString(token.String())
			}
		case parser.CommentToken:
			// We ignore comments by default
		case parser.DoctypeToken:
			// We ignore doctypes by default - html5 does not require them and this is intended for sanitizing snippets of text
		default:
			// We ignore unknown token types by default

		}

	}

}

// cleanAttributes removes all attributes except those in the allowed list
func cleanAttributes(a []parser.Attribute, allowed []string) []parser.Attribute {
	if len(a) == 0 {
		return a
	}

	var cleaned []parser.Attribute
	for _, attr := range a {
		if includes(allowed, attr.Key) {
			cleaned = append(cleaned, attr)
		}
	}
	return cleaned
}

// includes returns true if this array of strings contains string s
func includes(a []string, s string) bool {
	for _, as := range a {
		if as == s {
			return true
		}
	}
	return false
}
