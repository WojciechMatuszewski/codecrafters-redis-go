package redis

import "strings"

type CommandType string

const (
	Echo        CommandType = "echo"
	Ping        CommandType = "ping"
	Set         CommandType = "set"
	Get         CommandType = "get"
	Cfg         CommandType = "config"
	Info        CommandType = "info"
	Pong        CommandType = "pong"
	ReplicaConf CommandType = "replconf"
	PSync       CommandType = "psync"
)

type Command struct {
	Type CommandType
	Args []string
}

func NewCommand(value Value) Command {
	switch value.Type {
	case Array:
		mType := strings.ToLower(value.Array[0].Bulk)
		var args []string
		for i := 1; i < len(value.Array); i++ {
			args = append(args, value.Array[i].Bulk)
		}

		return Command{Type: CommandType(mType), Args: args}

	case SimpleString:
		mType := strings.ToLower(value.SimpleString)
		return Command{Type: CommandType(mType), Args: []string{value.SimpleString}}

	case Bulk:
		mType := strings.ToLower(value.Bulk)
		return Command{Type: CommandType(mType), Args: []string{value.Bulk}}

	default:
		panic("Unknown value")
	}
}
