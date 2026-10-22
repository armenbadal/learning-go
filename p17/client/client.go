package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	pb "etm/gen/etasks"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	client := pb.NewTaskManagerClient(connection)

	ctxWatch, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()

	subscriptionRequest := &pb.SubscribeRequest{
		Name: *name,
	}
	stream, err := client.WatchTaskEvents(ctxWatch, subscriptionRequest)
	if err != nil {
		log.Fatalf("cannod subscribe: %v", err)
	}
	log.Println("Subscribed.")

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Println("Server closed the stream")
				break
			}
			if err != nil {
				log.Fatalf("failed to get message: %v", err)
			}

			log.Printf("--> [%s] %s %d\n", msg.GetTimestamp().AsTime(), msg.GetTaskId(), msg.GetType())
		}
	}()

	_, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for {
		var command string
		fmt.Scanf("%s", &command)
		if command == "exit" {
			break
		}
		fmt.Printf(">> %s\n", command)
	}
	// - - - -

	// rq := &pb.CreateTaskRequest{
	// 	Title:       "First",
	// 	Description: "My first task",
	// }
	// r, err := client.CreateTask(ctx, rq)
	// if err != nil {
	// 	log.Fatalf("CreateTask failed")
	// }
	// log.Printf("Task created with ID: %s", r.GetId())

	// rq = &pb.CreateTaskRequest{
	// 	Title:       "Second",
	// 	Description: "My second task",
	// }
	// r, err = client.CreateTask(ctx, rq)
	// if err != nil {
	// 	log.Fatalf("CreateTask failed")
	// }
	// log.Printf("Task created with ID: %s", r.GetId())

	// r2, err := client.GetTask(ctx, &pb.TaskIDRequest{Id: "TASK2"})
	// if err != nil {
	// 	log.Fatalf("Failed to get task: %v", err)
	// }
	// log.Printf("Task: <%s, %s, %s, %t>\n", r2.GetId(), r2.GetTitle(), r2.GetDescription(), r2.GetCompleted())

	// r3, err := client.ListTasks(ctx, &emptypb.Empty{})
	// if err != nil {
	// 	log.Fatalf("Failed to get the list of tasks: %v", err)
	// }

	// for _, task := range r3.Tasks {
	// 	log.Printf("Task: <%s, %s, %s, %t>\n", task.GetId(), task.GetTitle(), task.GetDescription(), task.GetCompleted())
	// }
}
