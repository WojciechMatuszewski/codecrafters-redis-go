package tcp

import (
	"fmt"
	"net"
)

const (
	protocol = "tcp"
)

type Server struct {
	address string
}

func NewServer(address string) *Server {
	return &Server{address: address}
}

type HandlerFunc func(conn net.Conn)

func (s *Server) ListenAndServe(handler HandlerFunc) error {
	l, err := net.Listen(protocol, s.address)
	if err != nil {
		return fmt.Errorf("failed to listen to %s connections on %s", protocol, s.address)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connections on %s", s.address)
		}

		go handleConnection(handler, conn)
	}
}

func handleConnection(handler HandlerFunc, conn net.Conn) {
	defer conn.Close()
	handler(conn)
}
