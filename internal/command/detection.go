package command

import (
	"bytes"
	"os"
)

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
	return plainTextType
}

func textLocale() string {
	for _, variable := range [...]string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if value := os.Getenv(variable); value != "" {
			return value
		}
	}
	return "C"
}
