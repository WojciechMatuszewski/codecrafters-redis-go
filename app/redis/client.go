package redis

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
)

type Client struct {
	store Store
}

func NewClient(store Store) *Client {
	return &Client{store: store}
}

type ClientInfo struct {
	Role       string
	ReplId     string
	ReplOffset string
}

func (c *Client) Handle(ctx context.Context, rw io.ReadWriter, info ClientInfo) {
	buf := make([]byte, 1024)

	n, err := rw.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}

		panic(err)
	}

	input := buf[:n]

	fmt.Printf("Received input: %q\n", string(input))

	message := ParseMessage(input)

	switch message.Type {
	case Ping:
		err := WriteSimpleString(rw, "PONG")
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}
	case Echo:
		arg := message.Values[0]
		err := WriteBulkString(rw, arg)
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}
	case Set:
		key := message.Values[0]
		value := message.Values[1]

		var expiry *int
		if len(message.Values) > 3 {
			rawExpiryMs := message.Values[3]
			expiryMs, err := strconv.Atoi(rawExpiryMs)
			if err != nil {
				log.Fatalln("Could not convert the expiry time to integer: %w", err)
			}

			expiry = &expiryMs
		}

		c.store.Set(key, value, expiry)

		err := WriteSimpleString(rw, "OK")
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}

	case Get:
		key := message.Values[0]
		value, found := c.store.Get(key)

		if !found {
			err := WriteNullBulkString(rw)
			if err != nil {
				log.Printf("Error handling %s command: %v", message.Type, err)
				return
			}

			return
		}

		err := WriteBulkString(rw, value)
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}

	case Info:
		err := WriteBulkString(rw, fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%s", info.Role, info.ReplId, info.ReplOffset))
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}

	case ReplicaConf:
		err := WriteSimpleString(rw, "OK")
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}

	case PSync:
		err := WriteSimpleString(rw, fmt.Sprintf("FULLRESYNC %s %s", info.ReplId, info.ReplOffset))
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}

		b64RDB := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
		data, err := base64.StdEncoding.DecodeString(b64RDB)
		if err != nil {
			log.Printf("Error decoding base64 RDB file %v, error: %v", message.Type, err)
			return
		}

		err = Write(rw, fmt.Sprintf("$%v\r\n%s", len(data), data))
		if err != nil {
			log.Printf("Error handling %s command: %v", message.Type, err)
			return
		}
	}
}
