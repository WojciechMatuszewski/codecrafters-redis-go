package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const address = "0.0.0.0:6379"

var state = map[string]string{}

func main() {
	server := redis.NewServer(address)
	server.ListenAndServe()
}
