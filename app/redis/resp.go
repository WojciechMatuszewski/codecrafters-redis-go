package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type RespType byte

const (
	rSimpleString RespType = '+'
	rError        RespType = '-'
	rInteger      RespType = ':'
	rBulk         RespType = '$'
	rArray        RespType = '*'
)

type ValueType string

const (
	Bulk         ValueType = "bulk"
	Array        ValueType = "array"
	SimpleString ValueType = "string"
)

type Value struct {
	Typ  ValueType
	Str  string
	Num  int
	Bulk string
	Arr  []Value
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(r io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(r)}
}

func (r *Resp) Read() (Value, error) {
	buf, err := r.reader.Peek(1)
	if err != nil {
		return Value{}, fmt.Errorf("failed to read byte: %w", err)
	}
	if len(buf) > 1 {
		return Value{}, errors.New("issue with peeking the buffer")
	}

	_type := RespType(buf[0])
	switch _type {
	case rArray:
		return r.readArray()
	case rBulk:
		return r.readBulk()
	case rSimpleString:
		return r.readSimpleString()
	default:
		fmt.Println("Unknown type")
	}

	return Value{}, nil
}

func (r *Resp) readLine() ([]byte, error) {
	buf, err := r.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}

	line := buf[:len(buf)-2]
	return []byte(line), err
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{Typ: Bulk}

	typeLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading bulk: %w", err)
	}

	len, err := r.parseInteger(typeLine[1])
	if err != nil {
		panic(err)
	}

	contentLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading bulk: %w", err)
	}

	bulk := contentLine[0:len]
	v.Bulk = string(bulk)
	return v, nil
}

func (r *Resp) readSimpleString() (Value, error) {
	v := Value{Typ: SimpleString}

	contentLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading string: %w", err)
	}

	v.Str = string(contentLine[1:])
	return v, nil
}

func (r *Resp) readArray() (Value, error) {
	v := Value{Typ: Array}

	typeLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading array: %w", err)
	}

	len, err := r.parseInteger(typeLine[1])
	if err != nil {
		panic(err)
	}

	arr := make([]Value, len)
	for i := 0; i < len; i++ {
		val, err := r.Read()
		if err != nil {
			panic(err)
		}

		arr[i] = val
	}

	v.Arr = arr
	return v, nil
}

func (r *Resp) parseInteger(input byte) (int, error) {
	n, err := strconv.Atoi(string(input))
	if err != nil {
		return 0, err
	}

	return n, nil
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
