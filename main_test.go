package main

import (
	"strings"
	"testing"
)

func TestRunRejectsInvalidCommandArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing command", args: nil, want: "usage:"},
		{name: "unknown command", args: []string{"scan"}, want: "unknown command"},
		{name: "missing program", args: []string{"analyze"}, want: "--program"},
		{name: "invalid zombie age", args: []string{"zombie", "--days", "0"}, want: "greater than zero"},
		{name: "negative zombie age", args: []string{"zombie", "--days", "-2"}, want: "greater than zero"},
		{name: "missing duplicate keywords", args: []string{"dupes"}, want: "--keywords"},
		{name: "unexpected positional argument", args: []string{"zombie", "extra"}, want: "positional"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.args)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(test.want)) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}
