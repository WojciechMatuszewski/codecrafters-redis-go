package redis_test

import (
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	t.Run("foo", func(t *testing.T) {
		input := "*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*\r\n"
		len := len([]byte(input))
		fmt.Println(len)

		// buf.Write([]byte{})
	})
}

// func TestClient(t *testing.T) {
// 	ctx := context.Background()
// 	clientInfo := redis.ClientInfo{Role: "master"}

// 	commandTests := map[string]struct {
// 		input  []byte
// 		output []byte
// 	}{
// 		"PING": {
// 			input:  []byte(redis.FormatArray(redis.FormatBulkString("PING"))),
// 			output: []byte(redis.FormatSimpleString("PONG")),
// 		},
// 		"ECHO - ignores all but the first argument": {
// 			input: []byte(redis.FormatArray(
// 				redis.FormatBulkString("ECHO"),
// 				redis.FormatBulkString("hello"),
// 				redis.FormatBulkString("world"),
// 			)),
// 			output: []byte(redis.FormatBulkString("hello")),
// 		},
// 		"ECHO - strawberry": {
// 			input:  []byte("*2\r\n$4\r\nECHO\r\n$10\r\nstrawberry\r\n"),
// 			output: []byte(redis.FormatBulkString("strawberry")),
// 		},
// 	}

// 	for name, test := range commandTests {
// 		t.Run(fmt.Sprintf("Command test: %s \n", name), func(t *testing.T) {
// 			client := redis.NewClient(redis.NewInMemoryStore())

// 			buf := &bytes.Buffer{}
// 			buf.Write(test.input)

// 			client.Handle(ctx, buf, clientInfo)
// 			output, err := io.ReadAll(buf)

// 			assert.NoError(t, err)
// 			assert.Equal(t, string(test.output), string(output))
// 		})
// 	}

// 	t.Run("SET without expiry", func(t *testing.T) {
// 		client := redis.NewClient(redis.NewInMemoryStore())
// 		buf := &bytes.Buffer{}

// 		key := "hello"
// 		value := "world"

// 		buf.Write([]byte(redis.FormatArray(
// 			redis.FormatBulkString("SET"),
// 			redis.FormatBulkString(key),
// 			redis.FormatBulkString(value),
// 		)))
// 		client.Handle(ctx, buf, clientInfo)
// 		{
// 			read, err := io.ReadAll(buf)
// 			assert.NoError(t, err)
// 			assert.Equal(t, []byte(redis.FormatSimpleString("OK")), read)
// 		}

// 		buf.Write([]byte(redis.FormatArray(
// 			redis.FormatBulkString("GET"),
// 			redis.FormatBulkString(key),
// 		)))
// 		client.Handle(ctx, buf, clientInfo)
// 		{
// 			read, err := io.ReadAll(buf)
// 			assert.NoError(t, err)
// 			assert.Equal(t, []byte(redis.FormatBulkString(value)), read)
// 		}
// 	})

// 	t.Run("SET with expiry", func(t *testing.T) {
// 		now := time.Now()

// 		expiresAt := now.Add(3 * 1000 * time.Millisecond)
// 		expiryMs := (expiresAt.Sub(now)).Milliseconds()

// 		calls := 0
// 		nower := func() time.Time {
// 			// Setting the value
// 			if calls == 0 {
// 				calls = calls + 1
// 				return now
// 			}

// 			durationExpiry := time.Duration(expiryMs) * time.Millisecond

// 			if calls == 1 {
// 				calls = calls + 1

// 				notExpiredDuration := durationExpiry - 1*time.Millisecond
// 				newNow := now.Add(notExpiredDuration)

// 				return newNow
// 			}

// 			expiredDuration := durationExpiry + 1*time.Millisecond
// 			newNow := now.Add(expiredDuration)

// 			return newNow
// 		}

// 		client := redis.NewClient(redis.NewInMemoryStore(redis.WithNower(nower)))
// 		buf := &bytes.Buffer{}

// 		key := "hello"
// 		value := "world"

// 		buf.Write([]byte(redis.FormatArray(
// 			redis.FormatBulkString("SET"),
// 			redis.FormatBulkString(key),
// 			redis.FormatBulkString(value),
// 			redis.FormatBulkString("px"),
// 			redis.FormatBulkString(fmt.Sprintf("%v", expiryMs)),
// 		)))
// 		client.Handle(ctx, buf, clientInfo)
// 		{
// 			read, err := io.ReadAll(buf)
// 			assert.NoError(t, err)
// 			assert.Equal(t, []byte(redis.FormatSimpleString("OK")), read)
// 		}

// 		buf.Write([]byte(redis.FormatArray(
// 			redis.FormatBulkString("GET"),
// 			redis.FormatBulkString(key),
// 		)))
// 		client.Handle(ctx, buf, clientInfo)
// 		{
// 			read, err := io.ReadAll(buf)
// 			assert.NoError(t, err)
// 			assert.Equal(t, []byte(redis.FormatBulkString(value)), read)
// 		}

// 		buf.Write([]byte(redis.FormatArray(
// 			redis.FormatBulkString("GET"),
// 			redis.FormatBulkString(key),
// 		)))
// 		client.Handle(ctx, buf, clientInfo)
// 		{
// 			read, err := io.ReadAll(buf)
// 			assert.NoError(t, err)
// 			assert.Equal(t, []byte(redis.FormatNullBulkString()), read)
// 		}

// 		buf.Write([]byte(redis.FormatArray(
// 			redis.FormatBulkString("GET"),
// 			redis.FormatBulkString(key),
// 		)))
// 		client.Handle(ctx, buf, clientInfo)
// 		{
// 			read, err := io.ReadAll(buf)
// 			assert.NoError(t, err)
// 			assert.Equal(t, []byte(redis.FormatNullBulkString()), read)
// 		}

// 	})
// }
