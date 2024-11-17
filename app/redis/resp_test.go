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
			expected: redis.Value{Typ: redis.Bulk, Bulk: "hi there"},
		},
		"array1": {
			input: redis.FormatArray(
				redis.FormatBulkString("hi"),
				redis.FormatBulkString("there"),
			),
			expected: redis.Value{Typ: redis.Array, Arr: []redis.Value{
				{Typ: redis.Bulk, Bulk: "hi"},
				{Typ: redis.Bulk, Bulk: "there"},
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
			expected: redis.Value{Typ: redis.Array, Arr: []redis.Value{
				{Typ: redis.Bulk, Bulk: "SET"},
				{Typ: redis.Bulk, Bulk: "key"},
				{Typ: redis.Bulk, Bulk: "value"},
				{Typ: redis.Bulk, Bulk: "px"},
				{Typ: redis.Bulk, Bulk: "3000"},
			}},
		},
		"array ECHO with long string": {
			input: redis.FormatArray(
				redis.FormatBulkString("ECHO"),
				redis.FormatBulkString("1234567890"),
			),
			expected: redis.Value{Typ: redis.Array, Arr: []redis.Value{
				{Typ: redis.Bulk, Bulk: "ECHO"},
				{Typ: redis.Bulk, Bulk: "1234567890"},
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
			expected: redis.Value{Typ: redis.Array, Arr: []redis.Value{
				{Typ: redis.Bulk, Bulk: "1"},
				{Typ: redis.Bulk, Bulk: "2"},
				{Typ: redis.Bulk, Bulk: "3"},
				{Typ: redis.Bulk, Bulk: "4"},
				{Typ: redis.Bulk, Bulk: "5"},
				{Typ: redis.Bulk, Bulk: "6"},
				{Typ: redis.Bulk, Bulk: "7"},
				{Typ: redis.Bulk, Bulk: "8"},
				{Typ: redis.Bulk, Bulk: "9"},
				{Typ: redis.Bulk, Bulk: "10"},
			}},
		},
		"string": {
			input:    redis.FormatSimpleString("hi there"),
			expected: redis.Value{Typ: redis.SimpleString, Str: "hi there"},
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
