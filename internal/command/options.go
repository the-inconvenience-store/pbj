package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

type singleString struct {
	value string
	set   bool
}

func (value *singleString) Set(input string) error {
	if value.set {
		return errors.New("option may only be supplied once")
	}
	value.value, value.set = input, true
	return nil
}

func (value *singleString) String() string { return value.value }
func (*singleString) Type() string         { return "string" }

type singleBool struct {
	value bool
	set   bool
}

func (value *singleBool) Set(input string) error {
	if value.set {
		return errors.New("option may only be supplied once")
	}
	parsed, err := strconv.ParseBool(input)
	if err != nil || !parsed {
		return errors.New("informational option must be true")
	}
	value.value, value.set = parsed, true
	return nil
}

func (value *singleBool) String() string { return strconv.FormatBool(value.value) }
func (*singleBool) Type() string         { return "bool" }

type options struct {
	types     []string
	files     []string
	plain     singleString
	board     singleString
	help      singleBool
	version   singleBool
	listTypes singleBool
}

const generalBoard = "general"

func newRoot(opts *options) *cobra.Command {
	root := &cobra.Command{
		Use:   "pbjcopy",
		Short: "Copy stdin or file references to the macOS clipboard",
		Long: "Copy stdin or file references to the macOS clipboard.\n\n" +
			"Automatically recognizes PNG, JPEG, TIFF, PDF, RTF, and EPS; other input\n" +
			"is plain text decoded using LC_ALL, LC_CTYPE, LANG, or the C locale.\n" +
			"Explicit types label unchanged bytes; they do not convert formats.\n" +
			"Native pbcopy spellings -help and -pboard are also accepted.\n" +
			"Informational options must be requested alone. See docs/pbjcopy.md.",
		Example: "  printf hello | pbjcopy\n" +
			"  printf '<b>hello</b>' | pbjcopy -t public.html -p hello\n" +
			"  pbjcopy -f report.pdf -f screenshot.png",
		Args:               cobra.NoArgs,
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableSuggestions: true,
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	flags := root.Flags()
	flags.StringArrayVarP(&opts.types, "type", "t", nil, "Publish input under this pasteboard type (repeatable)")
	flags.StringArrayVarP(&opts.files, "file", "f", nil, "Copy a file reference (repeatable; does not read stdin)")
	flags.VarP(&opts.plain, "plain", "p", "Additional plain-text fallback")
	opts.board.value = generalBoard
	flags.Var(&opts.board, "pboard", "Select general, ruler, find, or font pasteboard (native -pboard spelling accepted)")
	flags.VarP(&opts.help, "help", "h", "Show help")
	flags.VarP(&opts.version, "version", "v", "Show version")
	flags.VarP(&opts.listTypes, "list-types", "l", "List common pasteboard types")
	for _, name := range []string{"help", "version", "list-types"} {
		flags.Lookup(name).NoOptDefVal = "true"
	}
	return root
}

func (opts options) validate(root *cobra.Command) error {
	if err := cobra.NoArgs(root, root.Flags().Args()); err != nil {
		return err
	}
	if err := opts.validateModes(); err != nil {
		return err
	}
	switch opts.board.value {
	case generalBoard, "ruler", "find", "font":
	default:
		return fmt.Errorf("unknown pasteboard %q; choose general, ruler, find, or font", opts.board.value)
	}
	for _, typeName := range opts.types {
		if typeName == "" || !utf8.ValidString(typeName) || strings.ContainsRune(typeName, 0) {
			return errors.New("pasteboard types must be nonempty UTF-8 identifiers without NUL")
		}
	}
	if opts.plain.set && !utf8.ValidString(opts.plain.value) {
		return errors.New("plain-text fallback must be valid UTF-8")
	}
	return nil
}

func (opts options) validateModes() error {
	information := 0
	for _, selected := range []bool{opts.help.value, opts.version.value, opts.listTypes.value} {
		if selected {
			information++
		}
	}
	copyOptions := len(opts.types) != 0 || len(opts.files) != 0 || opts.plain.set || opts.board.set
	if information > 1 || (information == 1 && copyOptions) {
		return errors.New("help, version, and type listing must be requested alone")
	}
	if len(opts.files) != 0 && (len(opts.types) != 0 || opts.plain.set) {
		return errors.New("file references cannot be combined with types or a plain-text fallback")
	}
	return nil
}

// nativeArguments adapts pbcopy/pbpaste single-dash long options without
// altering option values, including values that themselves look like flags.
func nativeArguments(args []string) []string {
	normalized := append([]string(nil), args...)
	for index := 0; index < len(normalized); index++ {
		arg := normalized[index]
		if arg == "--" {
			break
		}
		switch {
		case arg == "-help":
			normalized[index] = "--help"
		case arg == "-pboard" || strings.HasPrefix(arg, "-pboard="):
			normalized[index] = "-" + arg
		case arg == "-Prefer" || strings.HasPrefix(arg, "-Prefer="):
			normalized[index] = "--prefer" + strings.TrimPrefix(arg, "-Prefer")
		}
		switch arg {
		case "-t", "--type", "-p", "--plain", "-f", "--file", "-pboard", "--pboard", "-Prefer", "--prefer", "--item":
			index++
		}
	}
	return normalized
}
