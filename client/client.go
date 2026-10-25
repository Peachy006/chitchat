package main

import (
	"bufio"
	pb "chitchat/grpc"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"unicode/utf8"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clock int64
	mu    sync.Mutex
)

func tick() int64 {
	mu.Lock()
	defer mu.Unlock()
	clock++
	return clock
}

func update(received int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	clock = max(clock, received) + 1
	return clock
}

func main() {
	name := flag.String("name", "NoNameUser", "your name")
	addr := flag.String("addr", "localhost:9000", "server address")
	flag.Parse()

	// GRPC connection
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server %s: %v", *addr, err)
	}
	defer conn.Close()

	// Create a new chat client
	client := pb.NewChatClient(conn)

	// Connect to the chat server
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to server %s: %v", *addr, err)
	}

	//Join the chat
	t := tick()
	if err := stream.Send(&pb.ChatMessage{Sender: *name, Time: t}); err != nil {
		log.Fatalf("Failed to send join message: %v", err)
	}

	log.Printf("Joining the chat as %s at time %d", *name, t)

	done := make(chan bool)
	// Receive messages from the server
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				close(done)
				return
			}
			t := tick()
			log.Printf("Received message from %s: %s at time %d", msg.Sender, msg.Text, t)
		}
	}()

	//Use the chat
	fmt.Println("Write something in chat. Type /leave to leave.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/leave" {
			break
		}
		if !utf8.ValidString(text) || utf8.RuneCountInString(text) > 128 {
			fmt.Println("Max 128 characters allowed.")
			continue
		}
		stream.Send(&pb.ChatMessage{Sender: *name, Text: text, Time: tick()})
		log.Printf("Sent message from %s: %s", *name, text)
	}

	//Leave the chat
	stream.CloseSend()
	log.Printf("Leaving the chat as %s at logical time %d", *name, t)
	<-done
}
