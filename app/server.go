package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
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

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			log.Fatalf("Error reading from %s: %v", conn.RemoteAddr(), err)
		}

		fmt.Println("--- Received ---")
		fmt.Println(string(command))
		fmt.Println("---")

		_, err = writer.WriteString("+PONG\r\n")
		if err != nil {
			log.Printf("Error writing to %s: %v", conn.RemoteAddr(), err)
			return
		}

		writer.Flush()
	}
}
