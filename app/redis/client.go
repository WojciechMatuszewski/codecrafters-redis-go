package redis

import (
	"fmt"
	"strconv"
)

type Client struct {
	store Store
}

func NewClient(store Store) *Client {
	return &Client{store: store}
}

func (c *Client) Handle(cmd Command) (Value, error) {
	switch cmd.Type {
	case Ping:
		return Value{Type: SimpleString, SimpleString: "PONG"}, nil
	case Echo:
		return Value{Type: Bulk, Bulk: cmd.Args[0]}, nil
	case Get:
		key := cmd.Args[0]
		value, found := c.store.Get(key)
		if !found {
			return Value{Type: NullBulk}, nil
		}
		return Value{Type: Bulk, Bulk: value}, nil

	case Set:
		key := cmd.Args[0]
		value := cmd.Args[1]

		var expiry *int
		if len(cmd.Args) > 3 {
			rawExpiryMs := cmd.Args[3]
			expiryMs, err := strconv.Atoi(rawExpiryMs)
			if err != nil {
				return Value{}, fmt.Errorf("failed to convert expiry time to milliseconds: %w", err)
			}

			expiry = &expiryMs
		}

		c.store.Set(key, value, expiry)

		return Value{Type: SimpleString, SimpleString: "OK"}, nil
	}

	return Value{}, fmt.Errorf("unknown command: %v", cmd)
}
