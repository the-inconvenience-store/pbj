package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

type pasteOptions struct {
	typeName  singleString
	board     singleString
	to        singleString
	prefer    singleString
	item      singleString
	index     int
	listTypes singleBool
	files     singleBool
	null      singleBool
	help      singleBool
	version   singleBool
}

// Paste reads the clipboard without modifying it. Like Run, it returns 0 for
// success, 1 for operational failure, and 2 for an invalid invocation.
// Stdin is never read. The caller supplies Stdout and Stderr.
// --to copies file references to the filesystem instead of writing stdout.
func (cmd Command) Paste(args []string) int {
	code := 2
	var opts pasteOptions
	root := newPasteRoot(&opts)
	root.RunE = func(root *cobra.Command, _ []string) error {
		if err := opts.validate(root); err != nil {
			return err
		}
		code = 1
		var data []byte
		var err error
		if opts.version.value {
			version := cmd.Version
			if version == "" {
				version = "dev"
			}
			data = fmt.Appendf(nil, "pbjpaste %s\n", version)
		} else {
			data, err = cmd.readPaste(opts)
		}
		if err != nil {
			return err
		}
		if opts.to.set {
			return nil
		}
		count, err := cmd.Stdout.Write(data)
		if err == nil && count != len(data) {
			return io.ErrShortWrite
		}
		return err
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
	root.SetOut(cmd.Stdout)
	root.SetErr(cmd.Stderr)
	executeError := root.Execute()
	if err := errors.Join(executeError, helpError); err != nil {
		if _, writeErr := fmt.Fprintf(cmd.Stderr, "pbjpaste: %v\n", err); writeErr != nil {
			return 1
		}
		return code
	}
	return 0
}

func newPasteRoot(opts *pasteOptions) *cobra.Command {
	root := &cobra.Command{
		Use:   "pbjpaste",
		Short: "Extract content from the macOS clipboard",
		Long: "Extract content from the macOS clipboard without changing it.\n\n" +
			"Default output prefers plain text, EPS, then RTF. Plain text uses\n" +
			"LC_ALL, LC_CTYPE, LANG, or C encoding; exact types preserve raw bytes.\n" +
			"Multiple items require --item for ordinary or exact extraction.\n" +
			"Use --to DIR to copy file references without overwriting existing names.\n" +
			"Native -help, -pboard, and -Prefer spellings are accepted.\n" +
			"Help and version must be requested alone. See docs/pbjpaste.md.",
		Example: "  pbjpaste --list-types\n" +
			"  pbjpaste -t public.png > image.png\n" +
			"  pbjpaste --files --null\n" +
			"  pbjpaste --to .",
		Args:               cobra.NoArgs,
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableSuggestions: true,
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.Flags().VarP(&opts.typeName, "type", "t", "Extract only this exact pasteboard type; never fall back")
	opts.board.value = generalBoard
	root.Flags().Var(&opts.board, "pboard", "Select general, ruler, find, or font pasteboard")
	opts.prefer.value = "txt"
	root.Flags().Var(&opts.prefer, "prefer", "Prefer txt, rtf, or ps; fall back to other supported formats")
	root.Flags().Var(&opts.item, "item", "Select a zero-based clipboard item")
	root.Flags().VarP(&opts.listTypes, "list-types", "l", "Inspect advertised types per item as JSON, without fetching data")
	root.Flags().Var(&opts.files, "files", "Extract ordered local file-reference paths as UTF-8 records")
	root.Flags().Var(&opts.to, "to", "Copy file references into this existing directory without overwriting")
	root.Flags().Var(&opts.null, "null", "Terminate paths with NUL instead of newline (requires --files)")
	root.Flags().VarP(&opts.help, "help", "h", "Show help")
	root.Flags().VarP(&opts.version, "version", "v", "Show version")
	for _, name := range []string{"list-types", "files", "null", "help", "version"} {
		root.Flags().Lookup(name).NoOptDefVal = "true"
	}
	return root
}

func (opts *pasteOptions) validate(root *cobra.Command) error {
	if err := cobra.NoArgs(root, root.Flags().Args()); err != nil {
		return err
	}
	if (opts.help.value || opts.version.value) && root.Flags().NFlag() != 1 {
		return errors.New("help and version must be requested alone")
	}
	switch opts.board.value {
	case generalBoard, "ruler", "find", "font":
	default:
		return fmt.Errorf("unknown pasteboard %q", opts.board.value)
	}
	switch opts.prefer.value {
	case "txt", "ascii", "rtf", "ps":
	default:
		return errors.New("preference must be txt, rtf, or ps")
	}
	if opts.item.set {
		index, err := strconv.Atoi(opts.item.value)
		if err != nil || index < 0 {
			return errors.New("item must be a nonnegative integer")
		}
		opts.index = index
	}
	return opts.validateExtraction()
}

func (opts *pasteOptions) validateExtraction() error {
	modes := 0
	for _, selected := range []bool{opts.listTypes.value, opts.files.value, opts.typeName.set, opts.prefer.set, opts.to.set} {
		if selected {
			modes++
		}
	}
	if modes > 1 {
		return errors.New("type inspection, exact type, files, preference, and filesystem paste are mutually exclusive")
	}
	if opts.null.value && !opts.files.value {
		return errors.New("--null requires --files")
	}
	if opts.typeName.set && (opts.typeName.value == "" || !utf8.ValidString(opts.typeName.value) || strings.ContainsRune(opts.typeName.value, 0)) {
		return errors.New("pasteboard type must be a nonempty UTF-8 identifier without NUL")
	}
	if opts.to.set && opts.to.value == "" {
		return errors.New("--to requires a nonempty destination directory")
	}
	return nil
}

func (cmd Command) readPaste(opts pasteOptions) (output []byte, err error) {
	open := cmd.OpenPasteboard
	if open == nil {
		open = pasteboard.Open
	}
	board, err := open(opts.board.value)
	if err != nil {
		return nil, fmt.Errorf("open pasteboard: %w", err)
	}
	defer func() { err = errors.Join(err, board.Close()) }()
	snapshot, err := board.Capture()
	if err != nil {
		return nil, err
	}
	defer snapshot.Close()
	return opts.readSnapshot(snapshot)
}

func (opts *pasteOptions) readSnapshot(snapshot *pasteboard.Snapshot) ([]byte, error) {
	first, last, err := opts.itemRange(len(snapshot.Types()))
	if err != nil {
		return nil, err
	}
	if opts.listTypes.value {
		return inspectTypes(snapshot.Types(), first, last)
	}
	if opts.to.set {
		return nil, pasteFilesTo(snapshot, first, last, opts.to.value)
	}
	if opts.files.value {
		separator := byte('\n')
		if opts.null.value {
			separator = 0
		}
		return extractFilePaths(snapshot, first, last, separator)
	}
	if first == last {
		if opts.typeName.set {
			return nil, errors.New("clipboard has no items")
		}
		return nil, nil
	}
	if last-first > 1 {
		return nil, errors.New("select an item with --item when extracting multiple items")
	}
	if opts.typeName.set {
		return snapshot.Data(first, opts.typeName.value)
	}
	return defaultPaste(snapshot, first, opts.prefer.value)
}

func defaultPaste(snapshot *pasteboard.Snapshot, item int, preference string) ([]byte, error) {
	priorities := []string{plainTextType, "com.adobe.encapsulated-postscript", "public.rtf"}
	switch preference {
	case "rtf":
		priorities = []string{priorities[2], priorities[0], priorities[1]}
	case "ps":
		priorities[0], priorities[1] = priorities[1], priorities[0]
	}
	var unavailable error
	for _, typeName := range priorities {
		if !slices.Contains(snapshot.Types()[item], typeName) {
			continue
		}
		data, err := snapshot.Data(item, typeName)
		if errors.Is(err, pasteboard.ErrRepresentationUnavailable) {
			unavailable = err
			continue
		}
		if err == nil && typeName == plainTextType {
			return pasteboard.EncodeText(data, textLocale())
		}
		return data, err
	}
	return nil, unavailable
}

func (opts *pasteOptions) itemRange(count int) (int, int, error) {
	if !opts.item.set {
		return 0, count, nil
	}
	if opts.index >= count {
		return 0, 0, fmt.Errorf("clipboard item %d does not exist", opts.index)
	}
	return opts.index, opts.index + 1, nil
}

func inspectTypes(types [][]string, first, last int) ([]byte, error) {
	type itemTypes struct {
		Item  int      `json:"item"`
		Types []string `json:"types"`
	}
	items := make([]itemTypes, 0, last-first)
	for i := first; i < last; i++ {
		items = append(items, itemTypes{Item: i, Types: types[i]})
	}
	data, err := json.Marshal(items)
	return append(data, '\n'), err
}
