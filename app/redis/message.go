package redis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type CommandType string

const (
	Echo CommandType = "echo"
	Ping CommandType = "ping"
	Set  CommandType = "set"
	Get  CommandType = "get"
	Cfg  CommandType = "config"
	Info CommandType = "info"
	Pong CommandType = "pong"

	Ignore CommandType = "ignore"
)

type MsgType string

const (
	Array        MsgType = "*"
	SimpleString MsgType = "+"
)

type Message struct {
	Type CommandType
	Args []string
}

func ParseMessage(buf []byte) Message {
	if len(buf) == 0 {
		log.Fatalf("Malformed input data: empty buffer")
	}

	var cmd = Message{}
	reader := bufio.NewReader(bytes.NewReader(buf))

	rawTypeLine, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalln("Could not read line", err)
	}

	typeLine := strings.TrimSpace(string(rawTypeLine))
	msgType := typeLine[0]
	if MsgType(msgType) == SimpleString {
		return Message{Type: Ignore, Args: []string{}}
	}

	cmdType, err := next(reader)
	if err != nil {
		log.Fatalln("Could not read line", err)
	}
	cmdType = strings.ToLower(cmdType)

	switch CommandType(cmdType) {
	case Ping:
		return Message{
			Type: Ping,
			Args: []string{},
		}
	case Echo:
		arg, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read args for echo", err)
		}

		return Message{
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
				return Message{Type: Set, Args: args}
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
		return Message{
			Type: Set,
			Args: append(args, expiry),
		}

	case Get:
		key, err := next(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}

		fmt.Println("GET command for key", key)

		return Message{
			Type: Get,
			Args: []string{key},
		}

	case Cfg:
		subCmd, err := next(reader)
		if err != nil {
			log.Fatalln("Could not read the CONFIG command sub-command", err)
		}

		if strings.ToLower(subCmd) != "get" {
			log.Fatalln("Unknown CONFIG sub-command", subCmd)
		}

		cfgKey, err := next(reader)
		if err != nil {
			log.Fatalln("Could not read CONFIG key", err)
		}

		return Message{
			Type: Cfg,
			Args: []string{subCmd, cfgKey},
		}

	case Info:
		subCmd, err := next(reader)
		if err != nil {
			log.Fatalln("Could not read the CONFIG command sub-command", err)
		}

		if strings.ToLower(subCmd) != "replication" {
			log.Fatalln("Unknown CONFIG sub-command", subCmd)
		}

		return Message{
			Type: Info,
			Args: []string{subCmd},
		}

	default:
		log.Fatalln("Unknown command", cmdType)
	}

	return cmd
}

func Write(w io.Writer, output string) error {
	fmt.Printf("Responding with: %q\n", output)
	_, err := w.Write([]byte(output))
	return err
}

func WriteBulkString(w io.Writer, input string) error {
	output := FormatBulkString(input)
	fmt.Printf("Responding with: %q\n", output)

	_, err := w.Write([]byte(output))
	return err
}

func WriteSimpleString(w io.Writer, input string) error {
	output := FormatSimpleString(input)
	fmt.Printf("Responding with: %q\n", output)

	_, err := w.Write([]byte(output))
	return err
}

func WriteNullBulkString(w io.Writer) error {
	output := FormatNullBulkString()
	fmt.Printf("Responding with: %q\n", output)

	_, err := w.Write([]byte(output))
	fmt.Println("Responding with null bulk")
	return err
}

func FormatBulkString(input string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(input), input)
}

func FormatSimpleString(input string) string {
	return fmt.Sprintf("+%s\r\n", input)
}

func FormatNullBulkString() string {
	return "$-1\r\n"
}

func FormatArray(elements ...string) string {
	output := fmt.Sprintf("*%v\r\n", len(elements))
	for _, element := range elements {
		output = fmt.Sprintf("%s%s", output, element)
	}

	return output
}

// const (
// 	typeBlobString   = '$'
// 	typeSimpleString = '+'
// 	typeArray        = '*'
// )

// type RESP struct {
// 	Type   byte
// 	values []string
// }

// func ReadMessage(reader *bufio.Reader) {
// 	typeLineBuf, err := reader.ReadBytes('\n')
// 	if err != nil {
// 		panic(err)
// 	}

// 	typeLine := strings.TrimSpace(string(typeLineBuf))
// 	mType := typeLine[0]
// 	switch mType {
// 	case typeArray:
// 		arrayLength, err := reader.ReadByte()
// 		if err != nil {
// 			panic(err)
// 		}

// 		arrayLength, err := strconv.Atoi(string(rawArrayLength))
// 		if err != nil {
// 			panic(err)
// 		}

// 		fmt.Println("arrayLength", arrayLength)

// 		out, err := io.ReadAll(reader)
// 		fmt.Println(string(out))
// 	}
// }

func next(reader *bufio.Reader) (string, error) {
	rawNextType, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read next: %w", err)
	}

	fmt.Println("rawNextType", string(rawNextType))

	nextType := strings.TrimSpace(string(rawNextType[0]))
	if nextType != "$" {
		return "", fmt.Errorf("unknown data type: %v", string(rawNextType))
	}

	rawNext, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read next: %w", err)
	}

	fmt.Println("rawNext", string(rawNext))

	next := strings.TrimSpace(string(rawNext))
	return next, nil
}
