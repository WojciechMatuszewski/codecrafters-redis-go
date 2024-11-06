package redis

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
	client  *Client
}

func NewServer(address string, config *Config) *Server {
	client := NewClient(NewInMemoryStore(), config)
	return &Server{address: address, client: client}
}

func (s *Server) ListenAndServe() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, protocol, s.address)
	if err != nil {
		return fmt.Errorf("failed to listen to %s connections on %s", protocol, s.address)
	}
	defer listener.Close()

	go listenLoop(ctx, listener, s.client)

	<-ctx.Done()

	fmt.Println("Shutting the server down")

	return nil
}

func listenLoop(ctx context.Context, listener net.Listener, client *Client) {
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
						client.Handle(conn)
					}
				}
			}()
		}
	}
}
