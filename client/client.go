package main

import (
	"log"
)

func main() {
	conn, err := grpc.NewClient("tcp", "localhost:9000")
	if err != nil {
		log.Fatal("Failed to listen on port 9000")

	}

}
