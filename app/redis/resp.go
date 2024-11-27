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
	NullBulk     ValueType = "nullBulk"
	Array        ValueType = "array"
	SimpleString ValueType = "string"
	Raw          ValueType = "raw"
)

type Value struct {
	Type         ValueType
	SimpleString string
	Number       int
	Bulk         string
	Array        []Value
	Raw          string
}

func (v Value) Format() string {
	switch v.Type {
	case Bulk:
		return FormatBulkString(v.Bulk)
	case NullBulk:
		return FormatNullBulkString()
	case Array:
		elements := make([]string, len(v.Array))
		for i := 0; i < len(v.Array); i++ {
			elements[i] = v.Array[i].Format()
		}
		return FormatArray(elements...)
	case SimpleString:
		return FormatSimpleString(v.SimpleString)
	case Raw:
		return v.Raw
	}

	panic("Unknown value type")
}

func (v Value) Write(w io.Writer) error {
	_, err := w.Write([]byte(v.Format()))
	return err
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
		return Value{}, errors.New("failed to read byte: buffer is empty")
	}

	_type := RespType(buf[0])
	fmt.Println("Read type", string(_type))
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
	var line []byte

	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	fmt.Printf("Read line %q\n", string(line))

	return line[:len(line)-2], nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{Type: Bulk}

	typeLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading bulk: %w", err)
	}

	len, err := r.parseInteger(typeLine[1:])
	if err != nil {
		panic(err)
	}

	contentLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading bulk: %w", err)
	}

	// fmt.Printf("Content line: %q\n", string(contentLine))
	bulk := contentLine[0:len]
	v.Bulk = string(bulk)
	return v, nil
}

func (r *Resp) readSimpleString() (Value, error) {
	v := Value{Type: SimpleString}

	contentLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading string: %w", err)
	}

	v.SimpleString = string(contentLine[1:])
	return v, nil
}

func (r *Resp) readArray() (Value, error) {
	v := Value{Type: Array}

	typeLine, err := r.readLine()
	if err != nil {
		return Value{}, fmt.Errorf("failed to read line while reading array: %w", err)
	}

	len, err := r.parseInteger(typeLine[1:])
	if err != nil {
		panic(err)
	}

	arr := make([]Value, len)
	for i := int64(0); i < len; i++ {
		val, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return v, nil
			}

			panic(err)
		}

		arr[i] = val
	}

	v.Array = arr
	return v, nil
}

func (r *Resp) parseInteger(input []byte) (int64, error) {
	n, err := strconv.ParseInt(string(input), 10, 64)
	if err != nil {
		return 0, err
	}

	return n, nil
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
