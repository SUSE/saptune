package actions

import (
	"bytes"
	"strings"
	"testing"
)

func TestElizaReplies(t *testing.T) {
	e := NewEliza()

	tests := []struct {
		input    string
		contains string
	}{
		{
			input:    "Hello doctor",
			contains: "problem",
		},
		{
			input:    "I am very unhappy with my system",
			contains: "your system",
		},
		{
			input:    "You are being very difficult",
			contains: "I am being very difficult",
		},
		{
			input:    "Do computers think",
			contains: "computers",
		},
		{
			input:    "Can you help me",
			contains: "help you",
		},
		{
			input:    "My father told me about this",
			contains: "family",
		},
		{
			input:    "I feel happy today",
			contains: "feel",
		},
		{
			input:    "I need more memory for my database",
			contains: "your database",
		},
		{
			input:    "I can't tune the parameters",
			contains: "can't tune the parameters",
		},
		{
			input:    "I had a strange dream last night",
			contains: "dream",
		},
		{
			input:    "I am sorry",
			contains: "apologize",
		},
		{
			input:    "bye",
			contains: "Goodbye",
		},
		{
			input:    "",
			contains: "Please say something",
		},
	}

	for _, tt := range tests {
		reply := e.Reply(tt.input)
		if !strings.Contains(reply, tt.contains) {
			t.Errorf("Reply(%q) = %q; expected to contain %q", tt.input, reply, tt.contains)
		}
	}
}

func TestElizaReflections(t *testing.T) {
	e := NewEliza()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "I am unhappy with my job",
			expected: "you are unhappy with your job",
		},
		{
			input:    "you were talking about yourself",
			expected: "I was talking about myself",
		},
	}

	for _, tt := range tests {
		got := e.reflect(tt.input)
		if got != tt.expected {
			t.Errorf("reflect(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRunElizaInteractive(t *testing.T) {
	in := strings.NewReader("Hello\nI feel tired\nbye\n")
	out := &bytes.Buffer{}

	RunEliza(in, out)

	outStr := out.String()
	if !strings.Contains(outStr, "How do you do") {
		t.Errorf("expected initial greeting in output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "ELIZA:") {
		t.Errorf("expected ELIZA prefix in output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Goodbye") {
		t.Errorf("expected farewell on bye, got: %s", outStr)
	}
}
