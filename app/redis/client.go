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
			return Value{Type: Bulk, Bulk: "-1"}, nil
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

// func (c *Client) Handle(ctx context.Context, message Message) (Message, error) {
// 	switch message.Type {
// 	case Ping:
// 		return Message{Type: MessageType(SimpleString), Values: []Value{}}
// 	}
// }

// Pass role here?
// func (c *Client) Handle(ctx context.Context, rw io.ReadWriter, replicator Replicator) {
// 	buf := make([]byte, 1024)
// 	n, err := rw.Read(buf)
// 	if err != nil {
// 		if errors.Is(err, io.EOF) {
// 			return
// 		}

// 		panic(err)
// 	}

// 	input := buf[:n]

// 	fmt.Printf("Received input: %q\n", string(input))

// 	message := ParseMessage(input)

// 	switch message.Type {
// 	case Ping:
// 		err := WriteSimpleString(rw, "PONG")
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}
// 	case Echo:
// 		arg := message.Values[0]
// 		err := WriteBulkString(rw, arg)
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}
// 	case Set:
// 		key := message.Values[0]
// 		value := message.Values[1]

// 		var expiry *int
// 		if len(message.Values) > 3 {
// 			rawExpiryMs := message.Values[3]
// 			expiryMs, err := strconv.Atoi(rawExpiryMs)
// 			if err != nil {
// 				log.Fatalln("Could not convert the expiry time to integer: %w", err)
// 			}

// 			expiry = &expiryMs
// 		}

// 		c.store.Set(key, value, expiry)

// 		replicator.Replicate(rw, message)

// 		// Only return when the role === "master"
// 		err := WriteSimpleString(rw, "OK")
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}

// 	case Get:
// 		key := message.Values[0]
// 		value, found := c.store.Get(key)

// 		if !found {
// 			err := WriteNullBulkString(rw)
// 			if err != nil {
// 				log.Printf("Error handling %s command: %v", message.Type, err)
// 				return
// 			}

// 			return
// 		}

// 		err := WriteBulkString(rw, value)
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}

// 		// case Info:
// 		// 	err := replicator.HandleInfo(rw)
// 		// 	if err != nil {
// 		// 		log.Printf("Error handling %s command: %v", message.Type, err)
// 		// 		return
// 		// 	}

// 		// case ReplicaConf:
// 		// 	err := replicator.HandleReplicaConf(rw)
// 		// 	if err != nil {
// 		// 		log.Printf("Error handling %s command: %v", message.Type, err)
// 		// 		return
// 		// 	}

// 		// case PSync:
// 		// 	err = replicator.HandlePSync(rw)
// 		// 	if err != nil {
// 		// 		log.Printf("Error handling %s command: %v", message.Type, err)
// 		// 		return
// 		// 	}
// 	}
// }
