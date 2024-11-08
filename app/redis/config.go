package redis

import "fmt"

type Config struct {
	dir      string
	filename string
}

func NewConfig(dir string, dbfilename string) *Config {
	fmt.Printf("Created config on dir: %s and dbfilename: %s\n", dir, dbfilename)

	return &Config{dir: dir, filename: dbfilename}
}
