package main

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
	"github.com/codecrafters-io/redis-starter-go/app/tcp"
)

const address = "0.0.0.0:6379"

var state = map[string]string{}

func main() {
	server := tcp.NewServer(address)
	redisClient := redis.NewClient()

	server.ListenAndServe(func(conn net.Conn) {
		redisClient.Handle(conn)
	})
}
