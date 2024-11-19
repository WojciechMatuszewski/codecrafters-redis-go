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

	err = s.replicator.Connect(ctx)
	if err != nil {
		panic(err)
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

	// go s.listenLoop(ctx, listener)
	// return nil
}

// func (s *Server) connect(ctx context.Context, address string) (net.Conn, error) {
// 	if address == "" {
// 		return nil, nil
// 	}

// 	fmt.Printf("Connecting to address: %s\n", address)

// 	dialer := net.Dialer{}
// 	conn, err := dialer.DialContext(ctx, "tcp", address)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to address: %s, %w", address, err)
// 	}

// 	return conn, nil

// 	// err = s.handleHandshake(ctx, conn)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// go s.handleLoop(ctx, conn)

// 	// return nil
// }

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

			switch cmd.Type {
			case Info:
			case ReplicaConf:
			case PSync:
				outValues, err := s.replicator.Handle(cmd)
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

				_, err = connection.Write([]byte(outValue.Format()))
				if err != nil {
					log.Fatalf("failed to respond to client command: %v", err)
				}
			}
		}
	}
}

// func (s *Server) handleLoop(ctx context.Context, conn net.Conn) {
// 	defer conn.Close()
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			fmt.Println("Context is done!")
// 			return
// 		default:
// 			s.client.Handle(ctx, conn, s.replicator)
// 		}
// 	}
// }

// func (s *Server) handleHandshake(ctx context.Context, conn net.Conn) error {
// 	err := Write(conn, FormatArray(
// 		FormatBulkString("PING"),
// 	))
// 	if err != nil {
// 		return err
// 	}

// 	s.client.Handle(ctx, conn, s.replicator)

// 	err = Write(conn, FormatArray(
// 		FormatBulkString("REPLCONF"),
// 		FormatBulkString("listening-port"),
// 		FormatBulkString(s.Port),
// 	))
// 	if err != nil {
// 		return err
// 	}

// 	s.client.Handle(ctx, conn, s.replicator)

// 	err = Write(conn, FormatArray(
// 		FormatBulkString("REPLCONF"),
// 		FormatBulkString("capa"),
// 		FormatBulkString("psync2"),
// 	))
// 	if err != nil {
// 		return err
// 	}

// 	s.client.Handle(ctx, conn, s.replicator)

// 	err = Write(conn, FormatArray(
// 		FormatBulkString("PSYNC"),
// 		FormatBulkString("?"),
// 		FormatBulkString("-1"),
// 	))
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
