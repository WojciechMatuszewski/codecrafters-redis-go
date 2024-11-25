package redis

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	protocol = "tcp"
)

type Server struct {
	Host string
	Port string

	MasterHost string
	MasterPort string

	client *Client
	slaves []net.Conn

	logger *log.Logger
}

func NewServer(client *Client, host string, masterHost string, port string, masterPort string) *Server {
	server := &Server{
		Host:       host,
		Port:       port,
		MasterHost: masterHost,
		MasterPort: masterPort,

		client: client,
		slaves: []net.Conn{},
	}
	logger := log.New(os.Stdout, fmt.Sprintf("[%s on %s:%s] ", server.role(), server.Host, server.Port), 0)
	server.logger = logger

	return server
}

func (s *Server) Address() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	s.logger.Print("Starting the server")

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, err := s.listen(ctx, s.Address())
	if err != nil {
		return fmt.Errorf("failed to listen on address: %s, %w", s.Address(), err)
	}
	go s.serve(ctx, listener)

	err = s.masterHandshake(ctx)
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
	s.logger.Println("Accepting connections")

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

			s.logger.Printf("New connection to the server: %s\n", connection.RemoteAddr())

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

			s.logger.Printf("Handling command: %q\n", cmd.value.Format())

			switch cmd.Type {
			case ReplConf:
				if cmd.Args[0] == "listening-port" {
					s.slaves = append(s.slaves, connection)
				}

				value := Value{Type: SimpleString, SimpleString: "OK"}
				err := value.Write(connection)
				if err != nil {
					fmt.Println("Failed to write", err)
				}
			case Info:
				info := fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%s", s.role(), "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", "0")
				value := Value{Type: Bulk, Bulk: info}
				err := value.Write(connection)
				if err != nil {
					fmt.Println("Failed to write", err)
				}
			case PSync:
				data := fmt.Sprintf("FULLRESYNC %s %s", "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb", "0")
				resyncValue := Value{
					Type:         SimpleString,
					SimpleString: data,
				}
				err := resyncValue.Write(connection)
				if err != nil {
					s.logger.Println("Failed to write", err)
				}

				b64RDB := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
				rdbData, err := base64.StdEncoding.DecodeString(b64RDB)
				if err != nil {
					return
				}
				rdbValue := Value{Type: Raw, Raw: fmt.Sprintf("$%v\r\n%s", len(rdbData), rdbData)}
				err = rdbValue.Write(connection)
				if err != nil {
					s.logger.Println("Failed to write", err)
				}

			default:
				outValue, err := s.client.Handle(cmd)
				if err != nil {
					s.logger.Fatalf("failed to handle client command: %v", err)
				}

				if err != nil {
					s.logger.Fatalf("failed to replicate command %v", cmd)
				}

				if s.role() == "master" {
					s.logger.Printf("Replicating: %q\n", cmd.value.Format())
					err := s.replicate(cmd)
					if err != nil {
						s.logger.Println("Failed to replicate", err)
					}
				}

				if s.role() == "replica" {
					s.logger.Println("Skipping the response")
					return
				}

				s.logger.Printf("Responding with: %q\n", outValue.Format())

				_, err = connection.Write([]byte(outValue.Format()))
				if err != nil {
					s.logger.Fatalf("failed to respond to client command: %v", err)
				}
			}

		}
	}
}

func (s *Server) masterHandshake(ctx context.Context) error {
	if s.MasterHost == "" || s.MasterPort == "" {
		return nil
	}

	address := fmt.Sprintf("%s:%s", s.MasterHost, s.MasterPort)
	s.logger.Printf("Connecting to address at: %s\n", address)

	dialer := net.Dialer{}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to address: %s, %w", address, err)
	}
	defer connection.Close()

	go s.handleLoop(ctx, connection)

	resp := NewResp(connection)
	{
		outValue := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "PING"},
		}}
		s.logger.Printf("Sending to master: %q\n", outValue.Format())

		err := outValue.Write(connection)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		resValue, err := resp.Read()
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		s.logger.Printf("Master responded with: %q\n", resValue.Format())
	}

	{

		outValue := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "REPLCONF"},
			{Type: Bulk, Bulk: "listening-port"},
			{Type: Bulk, Bulk: s.Port},
		}}
		s.logger.Printf("Sending to master: %q\n", outValue.Format())

		data := []byte(outValue.Format())
		_, err := connection.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		resValue, err := resp.Read()
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		s.logger.Printf("Master responded with: %q\n", resValue.Format())
	}

	{

		outValue := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "REPLCONF"},
			{Type: Bulk, Bulk: "capa"},
			{Type: Bulk, Bulk: "psync2"},
		}}
		s.logger.Printf("Sending to master: %q\n", outValue.Format())

		err = outValue.Write(connection)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		resValue, err := resp.Read()
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		s.logger.Printf("Master responded with: %q\n", resValue.Format())
	}

	{

		outValue := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "PSYNC"},
			{Type: Bulk, Bulk: "?"},
			{Type: Bulk, Bulk: "-1"},
		}}
		s.logger.Printf("Sending to master: %q\n", outValue.Format())

		err := outValue.Write(connection)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		resValue, err := resp.Read()
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		s.logger.Printf("Master responded with: %q\n", resValue.Format())
	}

	return nil
}

func (s *Server) role() string {
	if s.MasterHost == "" || s.MasterPort == "" {
		return "master"
	}

	return "slave"
}

func (s *Server) replicate(cmd Command) error {
	switch cmd.Type {
	case Set:
		fmt.Printf("Replicating %v command\n", cmd)
		for _, replica := range s.slaves {
			err := cmd.Write(replica)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}
