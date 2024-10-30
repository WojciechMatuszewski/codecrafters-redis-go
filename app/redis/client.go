package redis

import (
	"errors"
	"fmt"
	"io"
	"log"
)

type Client struct {
	store Store
}

func NewClient() *Client {
	return &Client{store: NewInMemoryStore()}
}

func (c *Client) Handle(rw io.ReadWriter) {
	buf := make([]byte, 1024)
	n, err := rw.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}

		panic(err)
	}

	input := buf[:n]
	command := ParseCommand(input)

	switch command.Type {
	case Ping:
		_, err := rw.Write([]byte("+PONG\r\n"))
		if err != nil {
			log.Printf("Error handling %s command: %v", command.Type, err)
			return
		}
	case Echo:
		arg := command.Args[0]
		output := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)

		_, err = rw.Write([]byte(output))
		if err != nil {
			log.Printf("Error handling %s command: %v", command.Type, err)
			return
		}
	case Set:
		key := command.Args[0]
		value := command.Args[1]

		c.store.Set(key, value)

		_, err = rw.Write([]byte("+OK\r\n"))
		if err != nil {
			log.Printf("Error handling %s command: %v", command.Type, err)
			return
		}

	case Get:
		key := command.Args[0]
		value, found := c.store.Get(key)

		if !found {
			_, err = rw.Write([]byte("$-1\r\n")) // Redis nil response
			if err != nil {
				log.Printf("Error handling %s command: %v", command.Type, err)
				return
			}

			return
		}

		output := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
		_, err = rw.Write([]byte(output))
		if err != nil {
			log.Printf("Error handling %s command: %v", command.Type, err)
			return
		}
	}
}
