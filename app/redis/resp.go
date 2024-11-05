package redis

import "fmt"

func BulkString(input string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(input), input)
}

func SimpleString(input string) string {
	return fmt.Sprintf("+%s\r\n", input)
}

func NullBulkString() string {
	return "$-1\r\n"
}

func Array(elements ...string) string {
	output := fmt.Sprintf("*%v\r\n", len(elements))
	for _, element := range elements {
		output = fmt.Sprintf("%s%s", output, element)
	}

	return output
}
