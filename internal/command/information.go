package command

import (
	"fmt"
	"io"
)

const commonTypes = `public.utf8-plain-text  Plain text
public.html             HTML
public.rtf              Rich Text Format
public.png              PNG image
public.jpeg             JPEG image
public.tiff             TIFF image
com.adobe.pdf           PDF document
com.adobe.encapsulated-postscript  Encapsulated PostScript
public.url              URL
public.file-url         File reference
public.json             JSON
public.xml              XML

This is a common-type list, not a restriction; custom identifiers are accepted.
`

func (cmd Command) information(opts options) error {
	if opts.version.value {
		version := cmd.Version
		if version == "" {
			version = "dev"
		}
		_, err := fmt.Fprintf(cmd.Stdout, "pbjcopy %s\n", version)
		return err
	}
	_, err := io.WriteString(cmd.Stdout, commonTypes)
	return err
}

// checkedOutput retains errors that Cobra's help renderer cannot return.
type checkedOutput struct {
	target io.Writer
	err    error
}

func (output *checkedOutput) Write(data []byte) (int, error) {
	count, err := output.target.Write(data)
	if output.err == nil {
		output.err = err
	}
	return count, err
}
