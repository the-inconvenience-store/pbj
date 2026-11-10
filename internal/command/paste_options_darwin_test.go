//go:build darwin && cgo

package command_test

import (
	"testing"
)

func TestPasteRejectsAmbiguousAndRepeatedOptions(t *testing.T) {
	cases := [][]string{
		{optionHelp, optionItem, "0"},
		{optionHelp, optionVersion},
		{optionVersion, "--pboard", "general"},
		{optionHelp, optionHelp},
		{"--help=false"},
		{optionListTypes, optionFiles},
		{optionListTypes, "--prefer", "txt"},
		{optionFiles, "-t", typePlain},
		{optionNull},
		{optionItem, "-1"},
		{optionItem, "0", optionItem, "1"},
		{"-t", typePlain, "--type", typeHTML},
		{"-t", ""},
		{"-t", string([]byte{255})},
		{"-t", "abc\x00def"},
		{"--", "operand"},
	}
	for _, args := range cases {
		t.Run(args[0]+" invocation", func(t *testing.T) {
			run := newInvocation(t, nil)
			if code := run.command.Paste(args); code != 2 || run.stdout.Len() != 0 {
				t.Fatalf("args=%q exit=%d stdout=%q stderr=%s", args, code, &run.stdout, &run.stderr)
			}
		})
	}
}
