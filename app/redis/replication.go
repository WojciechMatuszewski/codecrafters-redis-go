package redis

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
)

type replica struct {
	host       string
	port       string
	connection net.Conn
}

type ReplicationInfo struct {
	MasterHost string
	MasterPort string

	Host string
	Port string

	ReplId     string
	ReplOffset string
	Role       string
}

func (ri *ReplicationInfo) MasterAddress() string {
	if ri.MasterHost == "" || ri.MasterPort == "" {
		return ""
	}

	return fmt.Sprintf("%s:%s", ri.MasterHost, ri.MasterPort)
}

type Replicator interface {
	Handle(ctx context.Context, cmd Command) ([]Value, error)
	Replicate(ctx context.Context, cmd Command) error
	MasterHandshake(ctx context.Context) error
	Role() string
}

type ServerReplicator struct {
	info     *ReplicationInfo
	replicas []replica
}

func NewServerReplicator(host string, port string, replicaof string) Replicator {
	info := &ReplicationInfo{
		ReplOffset: "0",
		ReplId:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		MasterHost: "",
		MasterPort: "",
		Host:       host,
		Port:       port,
		Role:       "master",
	}

	if replicaof == "" {
		return &ServerReplicator{info: info, replicas: []replica{}}
	}

	addressParts := strings.Split(replicaof, " ")
	if len(addressParts) < 1 {
		return &ServerReplicator{info: info, replicas: []replica{}}
	}

	info.MasterHost = addressParts[0]
	info.MasterPort = addressParts[1]
	info.Role = "slave"

	return &ServerReplicator{info: info, replicas: []replica{}}
}

func (sr *ServerReplicator) Role() string {
	return sr.info.Role
}

func (sr *ServerReplicator) Replicate(ctx context.Context, cmd Command) error {
	switch cmd.Type {
	case Set:
		fmt.Printf("Replicating %v command\n", cmd)
		for _, replica := range sr.replicas {
			err := cmd.Write(replica.connection)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func (sr *ServerReplicator) Handle(ctx context.Context, cmd Command) ([]Value, error) {
	switch cmd.Type {
	case ReplConf:
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

func (sr *ServerReplicator) MasterHandshake(ctx context.Context) error {
	if sr.info.Role == "master" {
		return nil
	}

	address := sr.info.MasterAddress()
	connection, err := connect(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to connect to address: %s, %w", address, err)
	}
	defer connection.Close()

	{
		value := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "PING"},
		}}
		fmt.Printf("Sending to master: %q\n", value.Format())

		err := value.Write(connection)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		fmt.Printf("Master responded with: %q\n", string(buf[0:n]))
	}

	{

		value := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "REPLCONF"},
			{Type: Bulk, Bulk: "listening-port"},
			{Type: Bulk, Bulk: sr.info.Port},
		}}
		fmt.Printf("Sending to master: %q\n", value.Format())

		data := []byte(value.Format())
		_, err := connection.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		fmt.Println("Master responded with", string(buf[0:n]))
	}

	{

		value := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "REPLCONF"},
			{Type: Bulk, Bulk: "capa"},
			{Type: Bulk, Bulk: "psync2"},
		}}
		fmt.Printf("Sending to master: %q\n", value.Format())

		err = value.Write(connection)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		fmt.Println("Master responded with", string(buf[0:n]))
	}

	{

		value := Value{Type: Array, Array: []Value{
			{Type: Bulk, Bulk: "PSYNC"},
			{Type: Bulk, Bulk: "?"},
			{Type: Bulk, Bulk: "-1"},
		}}
		fmt.Printf("Sending to master: %q\n", value.Format())

		err := value.Write(connection)
		if err != nil {
			return fmt.Errorf("failed to write to master: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read %w", err)
		}

		fmt.Println("Master responded with", string(buf[0:n]))
	}

	return nil
}

func connect(ctx context.Context, address string) (net.Conn, error) {
	fmt.Printf("Connecting to address at: %s\n", address)

	dialer := net.Dialer{}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to address: %s, %w", address, err)
	}

	return connection, nil
}
