package redis

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"strings"
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

func (sc ServerConfig) ReplicaAddress() string {
	if sc.Replica == "" {
		return ""
	}

	addressParts := strings.Split(sc.Replica, " ")
	if len(addressParts) < 1 {
		return ""
	}

	return fmt.Sprintf("%s:%s", addressParts[0], addressParts[1])
}

type Server struct {
	config ServerConfig
	client *Client
}

func NewServer(config ServerConfig, client *Client) *Server {
	return &Server{config: config, client: client}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	fmt.Printf("Starting the server: %s\n", s.config.Address())

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := s.listen(ctx, s.config.Address())
	if err != nil {
		return err
	}

	err = s.connect(ctx, s.config.ReplicaAddress())
	if err != nil {
		return err
	}

	<-ctx.Done()

	fmt.Println("Shutting the server down")

	return nil
}

func (s *Server) listen(ctx context.Context, address string) error {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, protocol, address)
	if err != nil {
		return fmt.Errorf("failed to listen to %s connections on %s: %w", protocol, address, err)
	}

	go s.listenLoop(ctx, listener)
	return nil
}

func (s *Server) connect(ctx context.Context, address string) error {
	if address == "" {
		return nil
	}

	fmt.Printf("Connecting to address: %s\n", address)

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to address: %s, %w", address, err)
	}

	err = s.handleHandshake(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to establish the handshake: %w", err)
	}

	go s.handleLoop(ctx, conn)
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

				fmt.Println("error accepting the connection:", err)
				continue
			}

			fmt.Printf("New connection to the server: %s\n", conn.RemoteAddr())

			go s.handleLoop(ctx, conn)
		}
	}
}

func (s *Server) handleLoop(ctx context.Context, conn net.Conn) {
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
			s.client.Handle(ctx, conn, ClientInfo{Role: role, ReplId: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", ReplOffset: "0"})
		}
	}
}

func (s *Server) handleHandshake(ctx context.Context, conn net.Conn) error {
	err := Write(conn, FormatArray(
		FormatBulkString("PING"),
	))
	if err != nil {
		return err
	}

	s.client.Handle(ctx, conn, ClientInfo{Role: "slave"})

	err = Write(conn, FormatArray(
		FormatBulkString("REPLCONF"),
		FormatBulkString("listening-port"),
		FormatBulkString(s.config.Port),
	))
	if err != nil {
		return err
	}

	s.client.Handle(ctx, conn, ClientInfo{Role: "slave"})

	err = Write(conn, FormatArray(
		FormatBulkString("REPLCONF"),
		FormatBulkString("capa"),
		FormatBulkString("psync2"),
	))
	if err != nil {
		return err
	}

	s.client.Handle(ctx, conn, ClientInfo{Role: "slave"})

	err = Write(conn, FormatArray(
		FormatBulkString("PSYNC"),
		FormatBulkString("?"),
		FormatBulkString("-1"),
	))
	if err != nil {
		return err
	}

	return nil
}
