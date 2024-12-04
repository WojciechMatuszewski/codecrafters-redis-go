package redis

import (
	"io"
	"strings"
)

type CommandType string

const (
	Echo       CommandType = "echo"
	Ping       CommandType = "ping"
	Set        CommandType = "set"
	Get        CommandType = "get"
	Cfg        CommandType = "config"
	Info       CommandType = "info"
	Pong       CommandType = "pong"
	ReplConf   CommandType = "replconf"
	PSync      CommandType = "psync"
	Fullresync CommandType = "fullresync"
	Ok         CommandType = "ok"
	Wait       CommandType = "wait"
)

type Command struct {
	Type CommandType
	Args []string

	value Value
}

func (c Command) Write(w io.Writer) error {
	return c.value.Write(w)
}

func NewCommand(value Value) Command {
	switch value.Type {
	case Array:
		mType := strings.ToLower(value.Array[0].Bulk)
		var args []string
		for i := 1; i < len(value.Array); i++ {
			args = append(args, value.Array[i].Bulk)
		}
		return Command{Type: CommandType(mType), Args: args, value: value}

	case SimpleString:
		mType := strings.ToLower(value.SimpleString)
		return Command{Type: CommandType(mType), Args: []string{value.SimpleString}, value: value}

	case Bulk:
		mType := strings.ToLower(value.Bulk)
		return Command{Type: CommandType(mType), Args: []string{value.Bulk}, value: value}

	default:
		panic("Unknown value")
	}
}
