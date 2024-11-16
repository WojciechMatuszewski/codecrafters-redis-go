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
