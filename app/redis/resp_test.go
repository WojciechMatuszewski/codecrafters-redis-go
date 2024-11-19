package redis_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
	"github.com/stretchr/testify/assert"
)

func TestResp(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected redis.Value
	}{
		"bulk": {
			input:    redis.FormatBulkString("hi there"),
			expected: redis.Value{Type: redis.Bulk, Bulk: "hi there"},
		},
		"array1": {
			input: redis.FormatArray(
				redis.FormatBulkString("hi"),
				redis.FormatBulkString("there"),
			),
			expected: redis.Value{Type: redis.Array, Array: []redis.Value{
				{Type: redis.Bulk, Bulk: "hi"},
				{Type: redis.Bulk, Bulk: "there"},
			}},
		},
		"array2": {
			input: redis.FormatArray(
				redis.FormatBulkString("SET"),
				redis.FormatBulkString("key"),
				redis.FormatBulkString("value"),
				redis.FormatBulkString("px"),
				redis.FormatBulkString(fmt.Sprintf("%v", 3000)),
			),
			expected: redis.Value{Type: redis.Array, Array: []redis.Value{
				{Type: redis.Bulk, Bulk: "SET"},
				{Type: redis.Bulk, Bulk: "key"},
				{Type: redis.Bulk, Bulk: "value"},
				{Type: redis.Bulk, Bulk: "px"},
				{Type: redis.Bulk, Bulk: "3000"},
			}},
		},
		"array ECHO with long string": {
			input: redis.FormatArray(
				redis.FormatBulkString("ECHO"),
				redis.FormatBulkString("1234567890"),
			),
			expected: redis.Value{Type: redis.Array, Array: []redis.Value{
				{Type: redis.Bulk, Bulk: "ECHO"},
				{Type: redis.Bulk, Bulk: "1234567890"},
			}},
		},
		"big array": {
			input: redis.FormatArray(
				redis.FormatBulkString("1"),
				redis.FormatBulkString("2"),
				redis.FormatBulkString("3"),
				redis.FormatBulkString("4"),
				redis.FormatBulkString("5"),
				redis.FormatBulkString("6"),
				redis.FormatBulkString("7"),
				redis.FormatBulkString("8"),
				redis.FormatBulkString("9"),
				redis.FormatBulkString("10"),
			),
			expected: redis.Value{Type: redis.Array, Array: []redis.Value{
				{Type: redis.Bulk, Bulk: "1"},
				{Type: redis.Bulk, Bulk: "2"},
				{Type: redis.Bulk, Bulk: "3"},
				{Type: redis.Bulk, Bulk: "4"},
				{Type: redis.Bulk, Bulk: "5"},
				{Type: redis.Bulk, Bulk: "6"},
				{Type: redis.Bulk, Bulk: "7"},
				{Type: redis.Bulk, Bulk: "8"},
				{Type: redis.Bulk, Bulk: "9"},
				{Type: redis.Bulk, Bulk: "10"},
			}},
		},
		"string": {
			input:    redis.FormatSimpleString("hi there"),
			expected: redis.Value{Type: redis.SimpleString, SimpleString: "hi there"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := bytes.NewReader([]byte(test.input))
			resp := redis.NewResp(r)
			value, err := resp.Read()

			assert.NoError(t, err)
			assert.Equal(t, test.expected, value)
		})
	}
}
