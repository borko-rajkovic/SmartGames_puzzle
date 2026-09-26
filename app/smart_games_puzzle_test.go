package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestChooseBoard(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantPrompts int
	}{
		{
			name:        "flat board",
			input:       "1\n",
			wantName:    "Flat board",
			wantPrompts: 1,
		},
		{
			name:        "heart board",
			input:       "2\n",
			wantName:    "Heart-shaped board",
			wantPrompts: 1,
		},
		{
			name:        "retries invalid selection",
			input:       "unknown\n2\n",
			wantName:    "Heart-shaped board",
			wantPrompts: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			gotName, _, err := chooseBoard(strings.NewReader(test.input), &output)
			if err != nil {
				t.Fatalf("chooseBoard returned an error: %v", err)
			}
			if gotName != test.wantName {
				t.Errorf("chooseBoard name = %q; want %q", gotName, test.wantName)
			}
			if got := strings.Count(output.String(), "Choose a board to solve:"); got != test.wantPrompts {
				t.Errorf("printed %d prompts; want %d", got, test.wantPrompts)
			}
		})
	}
}

func TestChooseBoardReportsEndOfInput(t *testing.T) {
	_, _, err := chooseBoard(strings.NewReader(""), &bytes.Buffer{})
	if err == nil || err.Error() != "no board selection provided" {
		t.Fatalf("chooseBoard error = %v; want no board selection error", err)
	}
}
