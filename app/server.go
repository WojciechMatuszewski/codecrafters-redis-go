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

	buf := make([]byte, 1024)
	_, err := net.Conn.Read(conn, buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}

		panic(err)
	}

	fmt.Println("--- Received ---")
	fmt.Println(string(buf))
	fmt.Println("---")

	pong := []byte("+PONG\r\n")
	_, err = net.Conn.Write(conn, pong)
	if err != nil {
		panic(err)
	}
}
