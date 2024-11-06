package main

import (
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const address = "0.0.0.0:6379"

var state = map[string]string{}

func main() {

	config := redis.NewConfigFromArgs(os.Args)
	server := redis.NewServer(address, config)
	server.ListenAndServe()
}
