package main

import (
	pb "chitchat/grpc"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChatServer
	mu      sync.Mutex
	clock   int64
	clients map[string]chan *pb.ChatBroadcast
}

func (s *server) broadcast(sender, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock++
	msg := &pb.ChatBroadcast{Sender: sender, Text: text, Time: s.clock}
	for name, ch := range s.clients {
		if name == sender {
			continue
		}
		select {
		case ch <- msg:
		default:
			log.Printf("Client %s's channel is full, dropping message", name)
		}
	}
	log.Printf("Broadcasted message from %s: %d", sender, s.clock)
}

func (s *server) Connect(stream pb.Chat_ConnectServer) error {
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	sender := first.Sender
	ch := make(chan *pb.ChatBroadcast, 100)

	s.mu.Lock()
	s.clock = max(s.clock, first.Time) + 1
	s.clients[sender] = ch
	log.Printf("Client %s connected at logical time %d", sender, s.clock)
	s.mu.Unlock()
	s.broadcast(sender, "Client connected")

	go func() {
		for {
			msg, err := stream.Recv()
			s.mu.Lock()
			if err != nil {
				delete(s.clients, sender)
				close(ch)
				s.clock = s.clock + 1
				log.Printf("Client %s disconnected at logical time %d", sender, s.clock)
				s.broadcast(sender, "Client disconnected")
				s.mu.Unlock()
				return
			}
			s.clock = max(s.clock, msg.Time) + 1
			s.mu.Unlock()
			s.broadcast(sender, msg.Text)
		}
	}()

	for msg := range ch {
		stream.Send(msg)
		log.Printf("Sent message from %s: %s", msg.Sender, msg.Text)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal("Failed to listen on port 9000", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterChatServer(grpcServer, &server{clients: make(map[string]chan *pb.ChatBroadcast)})

	log.Printf("Server listening on port 9000")
	grpcServer.Serve(lis)

}
