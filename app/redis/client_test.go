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
			input:  []byte("*1\r\n$4\r\nPING\n"),
			output: []byte("+PONG\r\n"),
		},
		"ECHO - ignores all but the first argument": {
			input:  []byte("*3\r\n$4\r\nECHO\r\n$5\r\nhello\r\n$5\r\nworld\n"),
			output: []byte("$5\r\nhello\r\n"),
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

		buf.Write([]byte(fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%v\r\n%s\r\n$%v\r\n%s\n", len(key), key, len(value), value)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(string("+OK\r\n")), read)
		}

		buf.Write([]byte([]byte(fmt.Sprintf("*3\r\n$3\r\nGET\r\n$%v\r\n%s\n", len(key), key))))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(fmt.Sprintf("$%v\r\n%s\r\n", len(value), value)), read)
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

		buf.Write([]byte(fmt.Sprintf("*5\r\n$3\r\nset\r\n$%v\r\n%s\r\n$%v\r\n%s\r\n$2\r\npx\r\n$%v\r\n%v\r\n", len(key), key, len(value), value, len(fmt.Sprintf("%v", expiryMs)), expiryMs)))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(string("+OK\r\n")), read)
		}

		buf.Write([]byte([]byte(fmt.Sprintf("*3\r\n$3\r\nGET\r\n$%v\r\n%s\n", len(key), key))))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte(fmt.Sprintf("$%v\r\n%s\r\n", len(value), value)), read)
		}

		buf.Write([]byte([]byte(fmt.Sprintf("*3\r\n$3\r\nGET\r\n$%v\r\n%s\n", len(key), key))))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte("$-1\r\n"), read)
		}

		buf.Write([]byte([]byte(fmt.Sprintf("*3\r\n$3\r\nGET\r\n$%v\r\n%s\n", len(key), key))))
		client.Handle(buf)
		{
			read, err := io.ReadAll(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte("$-1\r\n"), read)
		}

	})
}
