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
	replica       = flag.String("replicaof", "", "replica of server")
)

func main() {
	flag.Parse()

	client := redis.NewClient(redis.NewInMemoryStore())

	server := redis.NewServer(redis.ServerConfig{
		Host:    defaultHost,
		Port:    *port,
		Replica: *replica,
	}, client)

	server.ListenAndServe()
}
