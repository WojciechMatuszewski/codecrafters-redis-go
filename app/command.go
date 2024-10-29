package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Type string

const (
	Echo = "echo"
	Ping = "ping"
	Set  = "set"
	Get  = "get"
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

	cmdType, err := readNext(reader)
	if err != nil {
		log.Fatalln("Could not read line", err)
	}
	cmdType = strings.ToLower(cmdType)

	switch cmdType {
	case Ping:
		return Command{
			Type: Ping,
			Args: []string{},
		}
	case Echo:
		arg, err := readNext(reader)
		if err != nil {
			log.Fatalln("Failed to read args for echo", err)
		}

		return Command{
			Type: Echo,
			Args: []string{arg},
		}
	case Set:
		key, err := readNext(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}

		value, err := readNext(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}

		return Command{
			Type: Set,
			Args: []string{key, value},
		}
	case Get:
		key, err := readNext(reader)
		if err != nil {
			log.Fatalln("Failed to read args for set", err)
		}

		return Command{
			Type: Get,
			Args: []string{key},
		}

	default:
		log.Fatalln("Unknown command", cmdType)
	}

	return cmd
}

func readNext(reader *bufio.Reader) (string, error) {
	rawNextType, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read next: %v", err)
	}
	nextType := strings.TrimSpace(string(rawNextType[0]))
	if nextType != "$" {
		return "", fmt.Errorf("unknown data type: %v", string(rawNextType))
	}

	rawNext, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read next: %v", err)
	}

	next := strings.TrimSpace(string(rawNext))
	return next, nil
}
