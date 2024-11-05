package redis_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	commandTests := map[string]struct {
		input  []byte
		output []byte
	}{
		"PING": {
			input:  []byte(redis.Array(redis.BulkString("PING"))),
			output: []byte(redis.SimpleString("PONG")),
		},
		"ECHO - ignores all but the first argument": {
			input: []byte(redis.Array(
				redis.BulkString("ECHO"),
				redis.BulkString("hello"),
				redis.BulkString("world"),
			)),
			output: []byte(redis.BulkString("hello")),
		},
	}

	for name, test := range commandTests {
		t.Run(fmt.Sprintf("Command test: %s \n", name), func(t *testing.T) {
			client := redis.NewClient(redis.NewInMemoryStore())

			buf := &bytes.Buffer{}
			buf.Write(test.input)

			client.Handle(buf)
			output, err := io.ReadAll(buf)

			assert.NoError(t, err)
			assert.Equal(t, test.output, output)
		})
	}

	t.Run("SET without expiry", func(t *testing.T) {
		client := redis.NewClient(redis.NewInMemoryStore())
		buf := &bytes.Buffer{}

		key := "hello"
		value := "world"

		buf.Write([]byte(redis.Array(
			redis.BulkString("SET"),
			redis.BulkString(key),
			redis.BulkString(value),
		)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(redis.SimpleString("OK")), read)
		}

		buf.Write([]byte(redis.Array(
			redis.BulkString("GET"),
			redis.BulkString(key),
		)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(redis.BulkString(value)), read)
		}
	})

	t.Run("SET with expiry", func(t *testing.T) {
		now := time.Now()

		expiresAt := now.Add(3 * 1000 * time.Millisecond)
		expiryMs := (expiresAt.Sub(now)).Milliseconds()

		calls := 0
		nower := func() time.Time {
			// Setting the value
			if calls == 0 {
				calls = calls + 1
				return now
			}

			durationExpiry := time.Duration(expiryMs) * time.Millisecond

			if calls == 1 {
				calls = calls + 1

				notExpiredDuration := durationExpiry - 1*time.Millisecond
				newNow := now.Add(notExpiredDuration)

				return newNow
			}

			expiredDuration := durationExpiry + 1*time.Millisecond
			newNow := now.Add(expiredDuration)

			return newNow
		}

		client := redis.NewClient(redis.NewInMemoryStore(redis.WithNower(nower)))
		buf := &bytes.Buffer{}

		key := "hello"
		value := "world"

		buf.Write([]byte(redis.Array(
			redis.BulkString("SET"),
			redis.BulkString(key),
			redis.BulkString(value),
			redis.BulkString("px"),
			redis.BulkString(fmt.Sprintf("%v", expiryMs)),
		)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(redis.SimpleString("OK")), read)
		}

		buf.Write([]byte(redis.Array(
			redis.BulkString("GET"),
			redis.BulkString(key),
		)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(redis.BulkString(value)), read)
		}

		buf.Write([]byte(redis.Array(
			redis.BulkString("GET"),
			redis.BulkString(key),
		)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(redis.NullBulkString()), read)
		}

		buf.Write([]byte(redis.Array(
			redis.BulkString("GET"),
			redis.BulkString(key),
		)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(redis.NullBulkString()), read)
		}

	})
}
