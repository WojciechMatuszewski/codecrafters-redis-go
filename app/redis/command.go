package redis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type CmdType string

const (
	Echo CmdType = "echo"
	Ping CmdType = "ping"
	Set  CmdType = "set"
	Get  CmdType = "get"
)

type Cmd struct {
	Type CmdType
	Args []string
}

func ParseCommand(buf []byte) Cmd {
	if len(buf) == 0 {
		log.Fatalf("Malformed input data: empty buffer")
	}

	var cmd = Cmd{}
	reader := bufio.NewReader(bytes.NewReader(buf))

	// Array
	rawSize, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalln("Could not read line", err)
	}

	_, err = strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(string(rawSize), "*", "")))
	if err != nil {
		log.Fatalln("Could not read the size of the array", err)
	}

	cmdType, err := next(reader)
	if err != nil {
		log.Fatalln("Could not read line", err)
	}
	cmdType = strings.ToLower(cmdType)

	switch CmdType(cmdType) {
	case Ping:
		return Cmd{
			Type: Ping,
			Args: []string{},
		}
	case Echo:
		arg, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read args for echo", err)
		}

		return Cmd{
			Type: Echo,
			Args: []string{arg},
		}
	case Set:
		key, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}

		value, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}
		args := []string{key, value}

		modifier, err := next(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Cmd{Type: Set, Args: args}
			}

			log.Fatalln("Failed to read args for set", err)
		}

		if modifier != "px" {
			log.Fatalln("Unknown SET modifier", modifier)
		}

		expiry, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read the expiry time for SET: %w", err)
		}
		return Cmd{
			Type: Set,
			Args: append(args, expiry),
		}

	case Get:
		key, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}

		return Cmd{
			Type: Get,
			Args: []string{key},
		}

	default:
		log.Fatalln("Unknown command", cmdType)
	}

	return cmd
}

func WriteBulkString(w io.Writer, input string) error {
	output := fmt.Sprintf("$%d\r\n%s\r\n", len(input), input)
	_, err := w.Write([]byte(output))
	return err
}

func WriteSimpleString(w io.Writer, input string) error {
	output := fmt.Sprintf("+%s\r\n", input)
	_, err := w.Write([]byte(output))
	return err
}

func WriteNullBulkString(w io.Writer) error {
	_, err := w.Write([]byte("$-1\r\n"))
	return err
}

func next(reader *bufio.Reader) (string, error) {
	rawNextType, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read next: %w", err)
	}

	nextType := strings.TrimSpace(string(rawNextType[0]))
	if nextType != "$" {
		return "", fmt.Errorf("unknown data type: %v", string(rawNextType))
	}

	rawNext, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read next: %w", err)
	}

	next := strings.TrimSpace(string(rawNext))
	return next, nil
}
