// Package command implements the pbjcopy and pbjpaste command-line interfaces.
package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

// Command owns one invocation's input, output, and pasteboard selection.
// OpenPasteboard may select an isolated native pasteboard for an embedded caller;
// nil selects the system pasteboard requested by the command line.
type Command struct {
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
	OpenPasteboard func(string) (*pasteboard.Board, error)
	Version        string
}

// Run copies input and returns a shell exit status: 0 success, 1 operational
// failure, or 2 invalid invocation. Streams must be provided by the caller.
func (cmd Command) Run(args []string) int {
	code := 2
	var opts options
	root := newRoot(&opts)
	root.RunE = func(root *cobra.Command, _ []string) error {
		if err := opts.validate(root); err != nil {
			return err
		}
		code = 1
		if opts.version.value || opts.listTypes.value {
			return cmd.information(opts)
		}
		return cmd.copyInput(opts)
	}
	var helpError error
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(root *cobra.Command, args []string) {
		helpError = opts.validate(root)
		if helpError != nil {
			return
		}
		code = 1
		output := &checkedOutput{target: cmd.Stdout}
		root.SetOut(output)
		defaultHelp(root, args)
		helpError = output.err
	})
	root.SetArgs(nativeArguments(args))
	root.SetIn(cmd.Stdin)
	root.SetOut(cmd.Stdout)
	root.SetErr(cmd.Stderr)
	if err := errors.Join(root.Execute(), helpError); err != nil {
		return cmd.report(code, err)
	}
	return 0
}

func (cmd Command) copyInput(opts options) error {
	var representations []pasteboard.Representation
	var files []string
	var err error
	switch {
	case opts.attachment.set:
		var path string
		path, err = cmd.prepareAttachment(opts.attachment.value)
		files = []string{path}
	case len(opts.files) != 0:
		files, err = normalizeFiles(opts.files)
	default:
		representations, err = cmd.prepareRepresentations(opts)
	}
	if err != nil {
		return err
	}
	open := cmd.OpenPasteboard
	if open == nil {
		open = pasteboard.Open
	}
	board, err := open(opts.board.value)
	if err != nil {
		if opts.attachment.set {
			err = errors.Join(err, os.RemoveAll(filepath.Dir(files[0])))
		}
		return fmt.Errorf("open pasteboard: %w", err)
	}
	if len(files) != 0 {
		err = board.WriteFiles(files)
	} else {
		err = board.Write(representations)
	}
	if err := errors.Join(err, board.Close()); err != nil {
		return fmt.Errorf("publish clipboard: %w", err)
	}
	return nil
}

const plainTextType = "public.utf8-plain-text"

func (cmd Command) prepareRepresentations(opts options) ([]pasteboard.Representation, error) {
	data, err := io.ReadAll(cmd.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	typeNames := opts.types
	if len(typeNames) == 0 {
		typeNames = []string{detectedType(data)}
		if typeNames[0] == plainTextType {
			decoded, err := pasteboard.DecodeText(data, textLocale())
			if err != nil {
				return nil, fmt.Errorf("decode plain text: %w", err)
			}
			data = decoded
		}
	}
	if opts.plain.set && slices.Contains(typeNames, plainTextType) {
		return nil, errors.New("plain-text fallback conflicts with the input's plain-text representation")
	}
	representations := make([]pasteboard.Representation, 0, len(typeNames)+1)
	seen := make(map[string]struct{}, len(typeNames))
	for _, typeName := range typeNames {
		if _, exists := seen[typeName]; !exists {
			representations = append(representations, pasteboard.Representation{Type: typeName, Data: data})
			seen[typeName] = struct{}{}
		}
	}
	if opts.plain.set {
		representations = append(representations, pasteboard.Representation{Type: plainTextType, Data: []byte(opts.plain.value)})
	}
	return representations, nil
}

func (cmd Command) report(code int, err error) int {
	if _, writeErr := fmt.Fprintf(cmd.Stderr, "pbjcopy: %v\n", err); writeErr != nil {
		return 1
	}
	return code
}
