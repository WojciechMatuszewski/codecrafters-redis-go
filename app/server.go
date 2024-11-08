package main

import (
	"flag"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const (
	defaultPort = "6379"
	defaultHost = "0.0.0.0"
)

var (
	cfgDir        = flag.String("dir", "", "config directory")
	cfgDbFilename = flag.String("dbfilename", "", "config dbfilename")
	port          = flag.String("port", defaultPort, "port of the server")
)

func main() {
	flag.Parse()

	config := redis.NewConfig(*cfgDir, *cfgDbFilename)
	server := redis.NewServer(defaultHost, *port, config)

	server.ListenAndServe()
}
