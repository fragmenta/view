package helpers

import (
	"bytes"
	"fmt"
	got "html/template"
	"io"
	"strings"

	parser "github.com/fragmenta/view/internal/html"
)

// HEAD

// We should handle a comma separated array of style names, as per rails

// Style inserts a css tag
func Style(name string) got.HTML {
	return got.HTML(fmt.Sprintf("<link href=\"/assets/styles/%s.css\" media=\"all\" rel=\"stylesheet\" type=\"text/css\" />", EscapeURL(name)))
}

// Script inserts a script tag
func Script(name string) got.HTML {
	return got.HTML(fmt.Sprintf("<script src=\"/assets/scripts/%s.js\" type=\"text/javascript\"></script>", EscapeURL(name)))
}

// HTML

// FIXME - REMOVE TWO NEXT METHODS - not required and ugly - better to use urls or html directly
// without reverse routing this is useless, and reverse routing is of dubious utility IMHO

// LinkTo is deprecated
func LinkTo(t string, f string, args ...interface{}) got.HTML {
	// We should always escape the args here...
	text := Escape(t)
	url := fmt.Sprintf(f, args...)
	return got.HTML(fmt.Sprintf("<a href=\"%s\">%s</a>", url, text))
}

// LinkToAttributes is deprecated
func LinkToAttributes(t string, url string, args ...string) got.HTML {
	// FIXME - this is a little messy, think of a better way to do this
	// we need to retain ability to feed in info to link formats via templates though
	// e.g. link_to "NAME", "url/%d/xxx", int
	text := Escape(t)
	attributes := ""
	if len(args) > 0 {
		attributes = strings.Join(args, " ")
	}
	return got.HTML(fmt.Sprintf("<a href=\"%s\" %s>%s</a>", url, attributes, text))
}

// Html returns s as a go template URL
func Html(s string) got.HTML {
	return got.HTML(s)
}

// HtmlAttribute formats a string as an HTMLAttr
func HtmlAttribute(s string) got.HTMLAttr {
	return got.HTMLAttr(s)
}

// Url returns s as a go template URL
func Url(s string) got.URL {
	return got.URL(s)
}

// Strip all html tags
func Strip(s string) got.HTML {
	return Sanitize(s, []string{}, []string{})
}

// Sanitize utf8 html, allowing some tags
// Usage: sanitize.HTMLAllowing("<b id=id>my html</b>",[]string{"b"},[]string{"id")
//- this uses the experimental html parser in golang external packages go.net
func Sanitize(s string, args ...[]string) got.HTML {

	var ignoreTags = []string{"title", "script", "style", "iframe", "frame", "frameset", "noframes", "noembed", "embed", "applet", "object"}
	var defaultTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "div", "span", "hr", "p", "br", "b", "i", "ol", "ul", "li", "strong", "em", "a", "img"}
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

func includes(a []string, s string) bool {
	for _, as := range a {
		if as == s {
			return true
		}
	}
	return false
}
