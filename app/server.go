package main

import (
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

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			log.Fatalf("Error reading from %s: %v", conn.RemoteAddr(), err)
		}

		command := Parse(buf[:n])
		switch command.Type {
		case Ping:
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				log.Printf("Error writing to %s: %v", conn.RemoteAddr(), err)
				return
			}
		case Echo:
			arg := command.Args[0]
			output := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
			fmt.Println(output)

			_, err = conn.Write([]byte(output))
			if err != nil {
				log.Printf("Error writing to %s: %v", conn.RemoteAddr(), err)
				return
			}
		}

		fmt.Println("--- Received ---")
		fmt.Println(string(buf[:n]))
		fmt.Println("---")

	}
}
