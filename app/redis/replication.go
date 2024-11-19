package redis

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
)

type ReplicationInfo struct {
	MasterHost string
	MasterPort string

	Host string
	Port string

	Replicas []string

	ReplId     string
	ReplOffset string
	Role       string
}

type Replicator interface {
	Connect(ctx context.Context) error
	Handle(cmd Command) ([]Value, error)
	Replicate(cmd Command) error
}

type ServerReplicator struct {
	info *ReplicationInfo
}

func NewServerReplicator(host string, port string, replicaof string) Replicator {
	info := &ReplicationInfo{
		ReplOffset: "0",
		ReplId:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		Replicas:   []string{},
		MasterHost: "",
		MasterPort: "",
		Host:       host,
		Port:       port,
		Role:       "master",
	}

	if replicaof == "" {
		return &ServerReplicator{info: info}
	}

	addressParts := strings.Split(replicaof, " ")
	if len(addressParts) < 1 {
		return &ServerReplicator{info: info}
	}

	info.MasterHost = addressParts[0]
	info.MasterPort = addressParts[1]
	info.Role = "slave"

	return &ServerReplicator{info: info}
}

func (sr *ServerReplicator) Replicate(cmd Command) error {
	return nil
}

func (sr *ServerReplicator) Handle(cmd Command) ([]Value, error) {
	switch cmd.Type {
	case ReplicaConf:
		if cmd.Args[0] == "listening-port" {
			sr.info.Replicas = append(sr.info.Replicas, fmt.Sprintf("%s:%s", sr.info.Host, cmd.Args[1]))
			return []Value{{Type: SimpleString, SimpleString: "OK"}}, nil
		}
		return []Value{{Type: SimpleString, SimpleString: "OK"}}, nil
	case Info:
		info := fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%s", sr.info.Role, sr.info.ReplId, sr.info.ReplOffset)
		return []Value{{Type: Bulk, Bulk: info}}, nil
	case PSync:
		data := fmt.Sprintf("FULLRESYNC %s %s", sr.info.ReplId, sr.info.ReplOffset)
		resyncValue := Value{
			Type:         SimpleString,
			SimpleString: data,
		}

		b64RDB := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
		rdbData, err := base64.StdEncoding.DecodeString(b64RDB)
		if err != nil {
			return []Value{}, fmt.Errorf("failed to parse b64RDB: %w", err)
		}
		rdbValue := Value{Type: Raw, Raw: fmt.Sprintf("$%v\r\n%s", len(rdbData), rdbData)}

		return []Value{resyncValue, rdbValue}, nil
	}

	return []Value{}, fmt.Errorf("unknown replication command: %v", cmd)
}

func (sr *ServerReplicator) Connect(ctx context.Context) error {
	if sr.info.MasterHost == "" || sr.info.MasterPort == "" {
		return nil
	}

	address := fmt.Sprintf("%s:%s", sr.info.MasterHost, sr.info.MasterPort)
	fmt.Printf("Connecting to address: %s\n", address)

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to address: %s, %w", address, err)
	}

	// err = s.handleHandshake(ctx, conn)
	// if err != nil {
	// 	return nil, err
	// }

	err = Write(conn, FormatArray(
		FormatBulkString("PING"),
	))
	if err != nil {
		return err
	}

	// client.Handle(ctx, conn, s.replicator)

	err = Write(conn, FormatArray(
		FormatBulkString("REPLCONF"),
		FormatBulkString("listening-port"),
		FormatBulkString(sr.info.Port),
	))
	if err != nil {
		return err
	}

	// client.Handle(ctx, conn, s.replicator)

	err = Write(conn, FormatArray(
		FormatBulkString("REPLCONF"),
		FormatBulkString("capa"),
		FormatBulkString("psync2"),
	))
	if err != nil {
		return err
	}

	// client.Handle(ctx, conn, s.replicator)

	err = Write(conn, FormatArray(
		FormatBulkString("PSYNC"),
		FormatBulkString("?"),
		FormatBulkString("-1"),
	))

	return err
}

// func (sr *ServerReplicator) Handle(ctx context.Context, rw io.ReadWriter) {
// 	buf := make([]byte, 1024)
// 	n, err := rw.Read(buf)
// 	if err != nil {
// 		if errors.Is(err, io.EOF) {
// 			return
// 		}

// 		panic(err)
// 	}
// 	input := buf[:n]

// 	fmt.Printf("Received input: %q\n", string(input))

// 	message := ParseMessage(input)
// 	switch message.Type {
// 	case Info:
// 		err := sr.handleInfo(rw)
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}

// 	case ReplicaConf:
// 		err := sr.handleReplicaConf(rw)
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}

// 	case PSync:
// 		err = sr.handlePSync(rw)
// 		if err != nil {
// 			log.Printf("Error handling %s command: %v", message.Type, err)
// 			return
// 		}
// 	}
// }

// func (sr *ServerReplicator) handleInfo(w io.Writer) error {
// 	err := WriteBulkString(w, fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%s", sr.info.Role, sr.info.ReplId, sr.info.ReplOffset))
// 	return err
// }

// func (sr *ServerReplicator) handleReplicaConf(w io.Writer) error {
// 	err := WriteSimpleString(w, "OK")
// 	return err
// }

// func (sr *ServerReplicator) HandleHandshake(ctx context.Context, conn net.Conn) error {
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
