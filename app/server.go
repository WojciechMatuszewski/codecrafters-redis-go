package main

import (
	"errors"
	"fmt"
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

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		fmt.Println("New connection:", conn.RemoteAddr())

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {

		buf := make([]byte, 1024)
		read, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			panic(err)
		}

		fmt.Println("--- Received ---")
		fmt.Println(string(buf[:read]))
		fmt.Println("---")

		pong := []byte("+PONG\r\n")
		_, err = conn.Write(pong)
		if err != nil {
			panic(err)
		}
	}
}
