package command

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"regexp"
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
	if xmlDeclaration.Match(text) && validXML(text) {
		return "public.xml"
	}
	return plainTextType
}

// validXML requires one complete root and no character data outside it.
// Token parsing does not resolve external entities or validate a DTD.
func validXML(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return roots == 1 && depth == 0
		}
		if err != nil {
			return false
		}
		switch value := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.TrimSpace(value)) != 0 {
				return false
			}
		}
	}
}

func textLocale() string {
	for _, variable := range [...]string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if value := os.Getenv(variable); value != "" {
			return value
		}
	}
	return "C"
}
