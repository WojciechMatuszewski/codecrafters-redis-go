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
	Cfg  CmdType = "config"
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

		fmt.Println("GET command for key", key)

		return Cmd{
			Type: Get,
			Args: []string{key},
		}

	case Cfg:
		subCmd, err := next(reader)
		if err != nil {
			log.Fatalln("Could not read the CONFIG command sub-command", err)
		}

		if subCmd != "get" {
			log.Fatalln("Unknown CONFIG sub-command", subCmd)
		}

		cfgKey, err := next(reader)
		if err != nil {
			log.Fatalln("Could not read CONFIG key", err)
		}

		return Cmd{
			Type: Cfg,
			Args: []string{subCmd, cfgKey},
		}

	default:
		log.Fatalln("Unknown command", cmdType)
	}

	return cmd
}

func Write(w io.Writer, output string) error {
	fmt.Println("Responding with", output)
	_, err := w.Write([]byte(output))
	return err
}

func WriteBulkString(w io.Writer, input string) error {
	output := BulkString(input)
	fmt.Println("Responding with", output)
	_, err := w.Write([]byte(output))
	return err
}

func WriteSimpleString(w io.Writer, input string) error {
	output := SimpleString(input)
	fmt.Println("Responding with", output)
	_, err := w.Write([]byte(output))
	return err
}

func WriteNullBulkString(w io.Writer) error {
	output := NullBulkString()
	fmt.Println("Responding with", output)
	_, err := w.Write([]byte(output))
	fmt.Println("Responding with null bulk")
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
