package command

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

const xmlHeader = `<\?xml[ \t\r\n]+version[ \t\r\n]*=[ \t\r\n]*(?:"1\.0"|'1\.0')(?:[ \t\r\n]+encoding[ \t\r\n]*=[ \t\r\n]*(?:"(?i:utf-8)"|'(?i:utf-8)'))?(?:[ \t\r\n]+standalone[ \t\r\n]*=[ \t\r\n]*(?:"(?:yes|no)"|'(?:yes|no)'))?[ \t\r\n]*\?>`

var xmlDeclaration = regexp.MustCompile(`^` + xmlHeader)

// HTML is deliberately a document-envelope heuristic, not a markup validator.
var htmlDocument = regexp.MustCompile(`(?s)^(?:` + xmlHeader + `[ \t\r\n]*)?(?i:(?:<!doctype[ \t\r\n]+html[ \t\r\n]*>[ \t\r\n]*)?<html(?:[ \t\r\n]+[^<>]*)?>.*</html[ \t\r\n]*>)$`)

func detectedType(data []byte) string {
	signatures := [...]struct {
		prefix   string
		typeName string
	}{
		{prefix: "\x89PNG", typeName: "public.png"},
		{prefix: "\xff\xd8\xff", typeName: "public.jpeg"},
		{prefix: "II\x2a\x00", typeName: "public.tiff"},
		{prefix: "MM\x00\x2a", typeName: "public.tiff"},
		{prefix: "%PDF", typeName: "com.adobe.pdf"},
		{prefix: "{\\rtf", typeName: "public.rtf"},
		{prefix: "%!PS-Adobe-", typeName: "com.adobe.encapsulated-postscript"},
	}
	for _, signature := range signatures {
		if bytes.HasPrefix(data, []byte(signature.prefix)) {
			return signature.typeName
		}
	}
	if !utf8.Valid(data) {
		return plainTextType
	}
	text := bytes.TrimSpace(data)
	if len(text) != 0 && (text[0] == '{' || text[0] == '[') && json.Valid(data) {
		return "public.json"
	}
	if htmlDocument.Match(text) {
		return "public.html"
	}
	if declaration := xmlDeclaration.FindIndex(text); declaration != nil && validXML(text[declaration[1]:]) {
		return "public.xml"
	}
	return plainTextType
}

// validXML requires one complete root and no character data outside it.
// Token parsing rejects directives rather than resolving external entities or DTDs.
func validXML(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return roots == 1 && depth == 0
		}
		if err != nil || !validXMLToken(token, depth) {
			return false
		}
		switch token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		}
	}
}

func validXMLToken(token xml.Token, depth int) bool {
	switch value := token.(type) {
	case xml.StartElement:
		// The token is consumed here; sorting its attributes needs no copy.
		slices.SortFunc(value.Attr, func(a, b xml.Attr) int {
			if order := strings.Compare(a.Name.Space, b.Name.Space); order != 0 {
				return order
			}
			return strings.Compare(a.Name.Local, b.Name.Local)
		})
		for i := 1; i < len(value.Attr); i++ {
			if value.Attr[i-1].Name == value.Attr[i].Name {
				return false
			}
		}
	case xml.Directive:
		return false
	case xml.ProcInst:
		return !strings.EqualFold(value.Target, "xml")
	case xml.CharData:
		return depth != 0 || len(bytes.TrimSpace(value)) == 0
	}
	return true
}

func textLocale() string {
	for _, variable := range [...]string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if value := os.Getenv(variable); value != "" {
			return value
		}
	}
	return "C"
}
