package tcp

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"syscall"
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

type ConnHandler func(conn net.Conn)

func (s *Server) ListenAndServe(handler ConnHandler) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, protocol, s.address)
	if err != nil {
		return fmt.Errorf("failed to listen to %s connections on %s", protocol, s.address)
	}
	defer l.Close()

	go handleListener(ctx, l, handler)

	<-ctx.Done()

	fmt.Println("Shutting down")

	return nil
}

func handleListener(ctx context.Context, listener net.Listener, handler ConnHandler) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Server closed. Not accepting any listeners anymore")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					fmt.Println("Listener closed. Stopping the accept loop")
					return
				}

				fmt.Println("Error accepting the connection:", err)
				continue
			}

			fmt.Println("New connection:", conn.RemoteAddr())

			go func() {
				defer conn.Close()
				for {
					select {
					case <-ctx.Done():
						fmt.Println("Context is done!")
						return
					default:
						handler(conn)
					}
				}
			}()
		}
	}
}

func handleConnection(handler ConnHandler, conn net.Conn) {
	// TODO: Read
	// https://trstringer.com/golang-deferred-function-error-handling/
	defer conn.Close()
	for {
		handler(conn)
	}
}
