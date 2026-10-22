package main

import (
	"context"
	"log"
	"time"

	pb "etm/gen/etasks"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	connection, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect: %v", err)
	}
	defer connection.Close()

	client := pb.NewTaskManagerClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rq := &pb.CreateTaskRequest{
		Title:       "First",
		Description: "My first task",
	}
	r, err := client.CreateTask(ctx, rq)
	if err != nil {
		log.Fatalf("CreateTask failed")
	}
	log.Printf("Task created with ID: %s", r.GetId())

	rq = &pb.CreateTaskRequest{
		Title:       "Second",
		Description: "My second task",
	}
	r, err = client.CreateTask(ctx, rq)
	if err != nil {
		log.Fatalf("CreateTask failed")
	}
	log.Printf("Task created with ID: %s", r.GetId())

	r2, err := client.GetTask(ctx, &pb.TaskIDRequest{Id: "TASK2"})
	if err != nil {
		log.Fatalf("Failed to get task: %v", err)
	}
	log.Printf("Task: <%s, %s, %s, %t>\n", r2.GetId(), r2.GetTitle(), r2.GetDescription(), r2.GetCompleted())

	r3, err := client.ListTasks(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("Failed to get the list of tasks: %v", err)
	}

	for _, task := range r3.Tasks {
		log.Printf("Task: <%s, %s, %s, %t>\n", task.GetId(), task.GetTitle(), task.GetDescription(), task.GetCompleted())
	}
}
