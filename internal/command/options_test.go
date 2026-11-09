package command_test

import (
	"bytes"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/the-inconvenience-store/pbj/internal/command"
)

func TestRejectsAmbiguousInvocations(t *testing.T) {
	cases := [][]string{
		{optionHelp, "-t", typeHTML},
		{optionHelp, "-help"},
		{optionVersion, optionListTypes},
		{optionBoard, "find", optionBoard, "general"},
		{optionPlain, "first", "-p", "second"},
		{optionFile, "some-file", optionType, typeHTML},
		{optionFile, "some-file", optionPlain, "text"},
		{optionBoard, "unknown"},
		{optionType, ""},
		{optionType, "bad\x00identifier"},
		{"--", optionFile},
	}
	for _, args := range cases {
		t.Run(args[0]+"/"+args[len(args)-1], func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := command.Command{
				Stdin:  iotest.ErrReader(errors.New("stdin was read")),
				Stdout: &stdout,
				Stderr: &stderr,
			}
			if code := cmd.Run(args); code != 2 {
				t.Fatalf("args=%q exit=%d, want usage failure 2; stderr=%s", args, code, &stderr)
			}
			if stdout.Len() != 0 {
				t.Fatalf("usage error emitted stdout: %q", &stdout)
			}
		})
	}
}
