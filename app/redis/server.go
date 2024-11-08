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
	host    string
	port    string
	address string
	client  *Client
}

func NewServer(host string, port string, config *Config) *Server {
	store := NewInMemoryStore()
	client := NewClient(store, config)

	address := fmt.Sprintf("%s:%s", host, port)
	return &Server{host: host, port: port, client: client, address: address}
}

func (s *Server) ListenAndServe() error {
	fmt.Printf("Starting the server on %s:%s\n", s.host, s.port)

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
