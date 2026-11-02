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

	_, err = client.RegisterMe(ctx, &wrapperspb.StringValue{Value: *name})
	if err != nil {
		log.Fatal("failed to register")
	}

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
		to, text, err := readInput(scanner)
		if err != nil {
			log.Println("invalid input format")
			continue
		}

		msg := &pb.TextToRequest{
			From: *name,
			To:   to,
			Text: text,
		}
		_, err = client.TextTo(ctx, msg)
		if err != nil {
			log.Printf("message sending failed: %v", err)
			continue
		}
	}
}

func readInput(scanner *bufio.Scanner) (string, string, error) {
	if scanner.Scan() {
		line := scanner.Text()
		from, text, found := strings.Cut(line, ">")
		if !found {
			return "", "", fmt.Errorf("invalid input format")
		}
		return from, text, scanner.Err()
	}
	return "", "", scanner.Err()
}
