package main

import (
	"context"
	"flag"
	"log"
	"strings"

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

	masterHost := ""
	masterPort := ""
	if *replicaof != "" {
		addressParts := strings.Split(*replicaof, " ")
		masterHost = addressParts[0]
		masterPort = addressParts[1]
	}

	client := redis.NewClient(redis.NewInMemoryStore())
	server := redis.NewServer(client, host, masterHost, *port, masterPort)
	err := server.ListenAndServe(context.Background())
	if err != nil {
		log.Fatalln("Server error:", err)
	}
}
