package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output redis.Cmd
	}{
		"PING": {
			input:  []byte(string("*1\r\n$4\r\nPING\n")),
			output: redis.Cmd{Type: redis.Ping, Args: []string{}},
		},
		"ECHO - ignores all but the first argument": {
			input:  []byte(string("*3\r\n$4\r\nECHO\r\n$5\r\nhello\r\n$5\r\nworld\n")),
			output: redis.Cmd{Type: redis.Echo, Args: []string{"hello"}},
		},
		"SET": {
			input:  []byte(string("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\n")),
			output: redis.Cmd{Type: redis.Set, Args: []string{"hello", "world"}},
		},
		"GET": {
			input:  []byte(string("*3\r\n$3\r\nGET\r\n$5\r\nhello\n")),
			output: redis.Cmd{Type: redis.Get, Args: []string{"hello"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := redis.ParseCommand(test.input)

			require.Equal(t, test.output, got)
		})
	}
}
