package redis

import "flag"

type Config struct {
	dir      string
	filename string
}

func NewConfig(dir string, dbfilename string) *Config {
	return &Config{dir: dir, filename: dbfilename}
}

func NewConfigFromArgs(args []string) *Config {
	flags := flag.NewFlagSet(args[0], flag.PanicOnError)

	dir := flags.String("dir", "", "Name of the directory")
	filename := flags.String("dbfilename", "", "Name of the file")

	err := flags.Parse(args[1:])
	if err != nil {
		panic(err)
	}

	return NewConfig(*dir, *filename)
}
