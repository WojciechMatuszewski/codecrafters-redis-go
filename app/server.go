package main

import (
	"context"
	"flag"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const (
	host        = "0.0.0.0"
	defaultPort = "6379"
)

var (
	port      = flag.String("port", defaultPort, "port of the server")
	replicaof = flag.String("replicaof", "", "is replica of")
)

func main() {
	flag.Parse()

	client := redis.NewClient(redis.NewInMemoryStore())
	replicator := redis.NewServerReplicator(host, *port, *replicaof)

	server := redis.NewServer(client, replicator, host, *port)
	err := server.ListenAndServe(context.Background())
	if err != nil {
		log.Fatalln("Server error:", err)
	}
}
