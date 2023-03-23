/*
add clickable anchors to headings and
add anchors to tablerows with a {#id}
MAKE headings unique use map to store current headings and then check if existsand if it does add -1 then check if that exists and if it does add -2
respect custom anchor with {#id}
*/
package main

import (
	"bytes"
	"fmt"
	"github.com/russross/blackfriday"
	"github.com/shurcooL/octicon"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strconv"
	"strings"
)

// a map containing all current anchors for the given file, to ensure unique IDs.
var anchors = make(map[string]bool)

// renders markdown
func Markdown(text []byte) []byte {
	const htmlFlags = 0
	renderer := &renderer{Html: blackfriday.HtmlRenderer(htmlFlags, "", "").(*blackfriday.Html)}
	unsanitized := blackfriday.Markdown(text, renderer, extensions)
	//THIS resets the anchors after each file gets processed
	// NASTY SOLUTION need better
	for k, _ := range anchors {
		delete(anchors, k)
	}
	return unsanitized
}

// Heading returns a heading HTML node with title text.
// The heading comes with an anchor based on the title.
//
// heading can be one of atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6.
func Heading(heading atom.Atom, title string) *html.Node {
	aName := blackfriday.SanitizedAnchorName(title)
	a := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{
			// Make this match appended string  on line 94
			{Key: atom.Name.String(), Val: aName},
			{Key: atom.Class.String(), Val: "headerlink"},
			{Key: atom.Href.String(), Val: "#" + aName},
			{Key: atom.Rel.String(), Val: "nofollow"},
			{Key: atom.Title.String(), Val: "Permanent link"},
			{Key: "aria-hidden", Val: "true"},
		},
	}
	span := &html.Node{
		// check what this does and get rid of it since it seems useless considering i got rid of span
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{{Key: atom.Class.String(), Val: "#"}},
		//remove this useless
		FirstChild: octicon.Link(),
	}
	a.AppendChild(span)
	h := &html.Node{Type: html.ElementNode, Data: heading.String()}
	h.AppendChild(a)
	h.AppendChild(&html.Node{Type: html.TextNode, Data: title})
	return h
}

// Extensions for the parser.
const extensions = blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
	blackfriday.EXTENSION_TABLES |
	blackfriday.EXTENSION_FENCED_CODE |
	blackfriday.EXTENSION_AUTOLINK |
	blackfriday.EXTENSION_STRIKETHROUGH |
	blackfriday.EXTENSION_SPACE_HEADERS |
	blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK

type renderer struct {
	*blackfriday.Html
}

// Table row anchors not checked for uniqueness or correctness
func (*renderer) TableRow(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	// nasty conversion, probably cleaner way of doing this

	contents, start, end := findId(string(text))

	inside := []rune(contents)
	//	fmt.Println("found possible tag")
	if len(inside)-1 > 0 && inside[0] == '#' {
		id := string(inside[1:])
		out.WriteString("<tr id=" + `"` + id + `"` + ">\n")
		newText := string(text[:start]) + string(text[end+1:])
		out.Write([]byte(newText))
	} else {
		out.WriteString("<tr>\n")
		out.Write(text)
	}
	out.WriteString("\n</tr>\n")
}

// finds {#id} and returns index for the brackets and the inside text
func findId(s string) (string, int, int) {
	i := strings.Index(s, "{")
	var j int
	if i >= 0 {
		j := strings.Index(s, "}")
		if j >= 0 {
			return s[i+1 : j], i, j
		}
	}
	return "", i, j
}

// Headings with clickable anchors.
func (*renderer) Header(out *bytes.Buffer, text func() bool, level int, _ string) {
	marker := out.Len()
	doubleSpace(out)

	if !text() {
		out.Truncate(marker)
		return
	}

	textHTML := out.String()[marker:]
	out.Truncate(marker)

	// Extract text content of the heading.
	var textContent string
	if node, err := html.Parse(strings.NewReader(textHTML)); err == nil {
		textContent = extractText(node)
	} else {
		// Failed to parse HTML (probably can never happen), so just use the whole thing.
		textContent = html.UnescapeString(textHTML)
	}
	anchorName := blackfriday.SanitizedAnchorName(textContent)
	original := anchorName + "-"
	var found int
	// this will loop until a unique anchor is found
	for anchors[anchorName] == true {
		found++
		anchorName = original + strconv.Itoa(found)
	}
	anchors[anchorName] = true

//	out.WriteString(fmt.Sprintf(`<h%d>%s<a id="%s" class="anchor" href="#%s" rel="nofollow" aria-hidden="true">#</a>`, level, textHTML, anchorName, anchorName))
		out.WriteString(fmt.Sprintf(`<h%d>%s<a id="%s" class="anchor" href="#%s" rel="nofollow" >#</a>`, level, textHTML, anchorName, anchorName))
	out.WriteString(fmt.Sprintf("</h%d>\n", level))
}

// extractText returns the recursive concatenation of the text content of an html node.
func extractText(n *html.Node) string {
	var out string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			out += c.Data
		} else {
			out += extractText(c)
		}
	}
	return out
}

// Unexported blackfriday helpers.

func doubleSpace(out *bytes.Buffer) {
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
}
