package redis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os/signal"
	"syscall"
)

const (
	protocol = "tcp"
)

type Server struct {
	Host       string
	Port       string
	client     *Client
	replicator Replicator
}

func NewServer(client *Client, replicator Replicator, host string, port string) *Server {
	return &Server{
		Host: host,
		Port: port,

		client:     client,
		replicator: replicator,
	}
}

func (s *Server) Address() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	fmt.Printf("Starting the server: %s\n", s.Address())

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, err := s.listen(ctx, s.Address())
	if err != nil {
		return fmt.Errorf("failed to listen on address: %s, %w", s.Address(), err)
	}
	go s.serve(ctx, listener)

	err = s.replicator.ConnectMaster(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	<-ctx.Done()

	fmt.Println("Shutting the server down")

	return nil
}

func (s *Server) listen(ctx context.Context, address string) (net.Listener, error) {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, protocol, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen to %s connections on %s: %w", protocol, address, err)
	}

	return listener, err

}

func (s *Server) serve(ctx context.Context, listener net.Listener) {
	fmt.Printf("Accepting connection on address: %s\n", listener.Addr())

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Server closed. Not accepting any listeners anymore")
			return
		default:
			connection, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					fmt.Println("Listener closed. Stopping the accept loop")
					return
				}

				fmt.Println("error accepting the connection:", err)
				continue
			}

			fmt.Printf("New connection to the server: %s\n", connection.RemoteAddr())

			go s.handleLoop(ctx, connection)

		}
	}
}

func (s *Server) handleLoop(ctx context.Context, connection net.Conn) {
	defer connection.Close()
	resp := NewResp(connection)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			value, err := resp.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				log.Fatalf("failed to handle connection: %v", err)
			}

			cmd := NewCommand(value)

			fmt.Println("Handling command", cmd)

			switch cmd.Type {
			case Info:
				outValues, err := s.replicator.Handle(ctx, cmd)
				if err != nil {
					log.Fatalf("failed to handle replicator command: %v", err)
				}

				for _, outValue := range outValues {
					_, err = connection.Write([]byte(outValue.Format()))
					if err != nil {
						log.Fatalf("failed to respond to client command: %v", err)
					}
				}
			case ReplConf:
				outValues, err := s.replicator.Handle(ctx, cmd)
				if err != nil {
					log.Fatalf("failed to handle replicator command: %v", err)
				}

				for _, outValue := range outValues {
					_, err = connection.Write([]byte(outValue.Format()))
					if err != nil {
						log.Fatalf("failed to respond to client command: %v", err)
					}
				}
			case PSync:
				outValues, err := s.replicator.Handle(ctx, cmd)
				if err != nil {
					log.Fatalf("failed to handle replicator command: %v", err)
				}

				for _, outValue := range outValues {
					_, err = connection.Write([]byte(outValue.Format()))
					if err != nil {
						log.Fatalf("failed to respond to client command: %v", err)
					}
				}

			default:
				outValue, err := s.client.Handle(cmd)
				if err != nil {
					log.Fatalf("failed to handle client command: %v", err)
				}

				err = s.replicator.Replicate(ctx, cmd)
				if err != nil {
					log.Fatalf("failed to replicate command %v", cmd)
				}

				if s.replicator.Role() == "slave" {
					fmt.Printf("Server is a replica. Skipping the response\n")
					return
				}

				fmt.Println("Responding with", outValue)

				_, err = connection.Write([]byte(outValue.Format()))
				if err != nil {
					log.Fatalf("failed to respond to client command: %v", err)
				}
			}
		}
	}
}
