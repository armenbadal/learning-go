package main

import (
	"context"
	pb "echat/gen/echat"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ChatRoom struct {
	mx sync.Mutex
	clients map[string]*participant
	pb.UnimplementedChatRoomServer
}

type participant struct {
	mx sync.Mutex
	messages chan *pb.TextFromResponse
}

func (cr *ChatRoom) RegisterMe(ctx context.Context, name *wrapperspb.StringValue) (*wrapperspb.BoolValue, error) {
	cr.mx.Lock()
	defer cr.mx.Unlock()

	cr.clients[name.Value]= &participant{
		messages: make(chan *pb.TextFromResponse, 4),
	}
	log.Printf("New participant registered: '%s'", name.Value)

	return &wrapperspb.BoolValue{Value: true}, nil
}

func (cr *ChatRoom) DeregisterMe(ctx context.Context, name *wrapperspb.StringValue) (*wrapperspb.BoolValue, error) {
	cr.mx.Lock()
	defer cr.mx.Unlock()

	delete(cr.clients, name.Value)

	return &wrapperspb.BoolValue{Value: true}, nil
}

func (cr *ChatRoom) TextTo(ctx context.Context, request *pb.TextToRequest) (*wrapperspb.BoolValue, error) {
	cr.mx.Lock()
	defer cr.mx.Unlock()

	from, to, text := request.GetFrom(), request.GetTo(), request.GetText()
	message := &pb.TextFromResponse{
		From: from,
		Text: text,
	}
	cr.clients[to].messages <- message

	return &wrapperspb.BoolValue{Value: true}, nil
}

func (cr *ChatRoom) TextFrom(name *wrapperspb.StringValue, stream pb.ChatRoom_TextFromServer) error {
	for {
		select {
		case event := <-cr.clients[name.Value].messages:
			if err := stream.Send(event); err != nil {
				log.Printf("failed to send task event to client %q: %v", name.Value, err)
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}

	return nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	cr := ChatRoom{
		clients: make(map[string]*participant),
	}
	pb.RegisterChatRoomServer(server, &cr)
	log.Printf("Chat room server listening at %v", listener.Addr())

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

