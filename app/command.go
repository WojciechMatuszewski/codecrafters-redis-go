package main

import (
	"bufio"
	"bytes"
	"log"
	"strconv"
	"strings"
)

type Type string

const (
	Echo = "echo"
	Ping = "ping"
)

type Command struct {
	Type Type
	Args []string
}

func Parse(buf []byte) Command {
	if len(buf) == 0 {
		log.Fatalf("Malformed input data: empty buffer")
	}

	var cmd = Command{}
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

	// Command type length
	_, err = reader.ReadBytes('\n')
	if err != nil {
		log.Fatalln("Could not read line", err)
	}

	// Command type
	rawCmdType, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalln("Could not read line", err)
	}

	cmdType := strings.ToLower(strings.TrimSpace(string(rawCmdType)))
	switch cmdType {
	case Ping:
		return Command{
			Type: Ping,
			Args: []string{},
		}
	case Echo:
		_, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalln("Failed to read args for echo", err)
		}

		rawArg, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalln("Failed to read args for echo", err)
		}

		arg := strings.TrimSpace(string(rawArg))
		return Command{
			Type: Echo,
			Args: []string{arg},
		}

	default:
		log.Fatalln("Unknown command", cmdType)
	}

	return cmd
}
