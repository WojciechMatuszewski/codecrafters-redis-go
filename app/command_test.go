package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output Command
	}{
		"PING": {
			input:  []byte(string("*1\r\n$4\r\nPING\n")),
			output: Command{Type: Ping, Args: []string{}},
		},
		"ECHO - ignores all but the first argument": {
			input:  []byte(string("*3\r\n$4\r\nECHO\r\n$5\r\nhello\r\n$5\r\nworld\n")),
			output: Command{Type: Echo, Args: []string{"hello"}},
		},
		"SET": {
			input:  []byte(string("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\n")),
			output: Command{Type: Set, Args: []string{"hello", "world"}},
		},
		"GET": {
			input:  []byte(string("*3\r\n$3\r\nGET\r\n$5\r\nhello\n")),
			output: Command{Type: Get, Args: []string{"hello"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := Parse(test.input)

			require.Equal(t, test.output, got)
		})
	}
}
