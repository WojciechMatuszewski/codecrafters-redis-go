package main

import (
	"errors"
	"io"
	"net"
)

const address = "0.0.0.0:6379"
const protocol = "tcp"

func main() {
	l, err := net.Listen(protocol, address)
	if err != nil {
		panic(err)
	}

	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		_, err = net.Conn.Read(conn, buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			panic(err)
		}

		pong := []byte("+PONG\r\n")
		_, err = net.Conn.Write(conn, pong)
		if err != nil {
			panic(err)
		}
	}

}
