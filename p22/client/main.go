package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	pb "echat/gen/echat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func main() {
	name := flag.String("name", "", "client name")
	flag.Parse()
	if strings.TrimSpace(*name) == "" {
		log.Fatal("the --name flag is required")
	}
	log.Printf("Starting client %q", *name)

	connection, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect: %v", err)
	}
	defer connection.Close()

	client := pb.NewChatRoomClient(connection)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := client.RegisterMe(ctx, &wrapperspb.StringValue{Value: *name})
	if err != nil {
		log.Fatal("failed to register")
	}
	fmt.Printf("|> %t\n", r.Value)

	req := &wrapperspb.StringValue{Value: *name}
	stream, err := client.TextFrom(ctx, req)
	if err != nil {
		log.Fatalf("cannot subscribe: %v", err)
	}

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			fmt.Printf("--> FROM: %s\nTEXT: %s\n", msg.GetFrom(), msg.GetText())
		}
	}()


	// sending loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		var to, text string
		fmt.Print("TO: ")
		if scanner.Scan() {
			to = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			continue
		}

		fmt.Print("TEXT: ")
		if scanner.Scan() {
			text = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			continue
		}

		msg := &pb.TextToRequest{
			From: *name,
			To: to,
			Text: text,
		}
		r, err := client.TextTo(ctx, msg)
		if err != nil {
			log.Println("message sending failed")
			continue
		}
		fmt.Printf("|> %t\n", r.Value)
	}
}
