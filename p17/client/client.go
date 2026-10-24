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
	"time"

	pb "etm/gen/etasks"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type command struct {
	action    string
	arguments []string
}

func parseCommand(text string) (*command, error) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid command: %q", text)
	}

	switch parts[0] {
	case "CREATE":
		if len(parts) < 3 {
			break
		}
		return &command{
			action:    "CREATE",
			arguments: []string{parts[1], strings.Join(parts[2:], " ")},
		}, nil
	case "GET", "COMPLETE", "DELETE":
		if len(parts) != 2 {
			break
		}
		return &command{action: parts[0], arguments: parts[1:]}, nil
	case "LIST":
		if len(parts) == 1 {
			return &command{action: "LIST"}, nil
		}
	}

	return nil, fmt.Errorf("invalid command: %q", text)
}

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

	client := pb.NewTaskManagerClient(connection)

	ctxWatch, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()

	subscriptionRequest := &pb.SubscribeRequest{
		Name: *name,
	}
	stream, err := client.WatchTaskEvents(ctxWatch, subscriptionRequest)
	if err != nil {
		log.Fatalf("cannot subscribe: %v", err)
	}
	log.Println("Subscribed.")

	type inputResult struct {
		line string
		err  error
	}
	input := make(chan inputResult, 1)
	go func() {
		defer close(input)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input <- inputResult{line: scanner.Text()}
		}
		if err := scanner.Err(); err != nil {
			input <- inputResult{err: err}
		}
	}()

	watchErr := make(chan error, 1)
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				watchErr <- io.EOF
				break
			}
			if err != nil {
				watchErr <- err
				break
			}

			log.Printf("--> [%s] %s %d\n", msg.GetTimestamp().AsTime(), msg.GetTaskId(), msg.GetType())
		}
	}()

	for {
		fmt.Print("> ")
		select {
		case err := <-watchErr:
			if err == io.EOF {
				log.Println("Server closed the stream")
			} else {
				log.Printf("failed to get message: %v", err)
			}
			return
		case result, ok := <-input:
			if !ok {
				return
			}
			if result.err != nil {
				log.Printf("failed to read command: %v", result.err)
				return
			}
			text := result.line
			if text == "exit" {
				return
			}

			cm, err := parseCommand(text)
			if err != nil {
				log.Printf("Invalid command: %v\n", err)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			switch cm.action {
			case "CREATE":
				rq := &pb.CreateTaskRequest{
					Title:       cm.arguments[0],
					Description: cm.arguments[1],
				}
				r, err := client.CreateTask(ctx, rq)
				if err != nil {
					log.Printf("CreateTask failed: %v", err)
					break
				}
				log.Printf("Task created with ID: %s", r.GetId())
			case "GET":
				rq := &pb.TaskIDRequest{Id: cm.arguments[0]}
				r, err := client.GetTask(ctx, rq)
				if err != nil {
					log.Printf("GetTask failed: %v", err)
					break
				}
				log.Printf("Task: <%s, %s, %s, %t>", r.GetId(), r.GetTitle(), r.GetDescription(), r.GetCompleted())
			case "LIST":
				r, err := client.ListTasks(ctx, &emptypb.Empty{})
				if err != nil {
					log.Printf("ListTasks failed: %v", err)
					break
				}
				for _, task := range r.GetTasks() {
					log.Printf("Task: <%s, %s, %s, %t>", task.GetId(), task.GetTitle(), task.GetDescription(), task.GetCompleted())
				}
			case "COMPLETE":
				rq := &pb.TaskIDRequest{Id: cm.arguments[0]}
				if _, err := client.CompleteTask(ctx, rq); err != nil {
					log.Printf("CompleteTask failed: %v", err)
					break
				}
				log.Printf("Task %s completed", cm.arguments[0])
			case "DELETE":
				rq := &pb.TaskIDRequest{Id: cm.arguments[0]}
				if _, err := client.DeleteTask(ctx, rq); err != nil {
					log.Printf("DeleteTask failed: %v", err)
					break
				}
				log.Printf("Task %s deleted", cm.arguments[0])
			}
			cancel()
		}
	}
}
