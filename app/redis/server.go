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

type ServerConfig struct {
	Host    string
	Port    string
	Replica string
}

func (sc ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", sc.Host, sc.Port)
}

type Server struct {
	config ServerConfig
	client *Client
}

func NewServer(config ServerConfig, client *Client) *Server {
	return &Server{config: config, client: client}
}

func (s *Server) ListenAndServe() error {
	fmt.Printf("Starting the server on %s\n", s.config.Address())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, protocol, s.config.Address())
	if err != nil {
		return fmt.Errorf("failed to listen to %s connections on %s", protocol, s.config.Address())
	}
	defer listener.Close()

	go s.listenLoop(ctx, listener)

	<-ctx.Done()

	fmt.Println("Shutting the server down")

	return nil
}

func (s *Server) listenLoop(ctx context.Context, listener net.Listener) {
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
						role := "master"
						if s.config.Replica != "" {
							role = "slave"
						}
						s.client.Handle(conn, ClientInfo{Role: role})
					}
				}
			}()
		}
	}
}
