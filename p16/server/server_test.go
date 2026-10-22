package main

import (
	"context"
	"testing"

	pb "etm/gen/etasks"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateTask(t *testing.T) {
	newTaskManager := func() *TaskManager {
		return &TaskManager{
			service: &TaskService{tasks: make(map[string]*Task)},
		}
	}

	t.Run("creates task", func(t *testing.T) {
		tm := newTaskManager()
		req := &pb.CreateTaskRequest{
			Title:       "Learn gRPC",
			Description: "Build a task manager",
		}

		got, err := tm.CreateTask(context.Background(), req)
		if err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
		if got.GetId() == "" {
			t.Fatal("CreateTask() returned an empty ID")
		}

		task, err := tm.service.Get(got.GetId())
		ok := err == nil
		if !ok {
			t.Fatalf("task with ID %q was not saved", got.GetId())
		}
		if task.Title != req.GetTitle() || task.Description != req.GetDescription() {
			t.Errorf("saved task = %+v; want title %q and description %q", task, req.GetTitle(), req.GetDescription())
		}
		if task.Completed {
			t.Error("new task should not be completed")
		}
	})

	t.Run("rejects nil request", func(t *testing.T) {
		tm := newTaskManager()

		_, err := tm.CreateTask(context.Background(), nil)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("CreateTask() error code = %v; want %v", status.Code(err), codes.InvalidArgument)
		}
	})

	t.Run("rejects empty title", func(t *testing.T) {
		tm := newTaskManager()

		_, err := tm.CreateTask(context.Background(), &pb.CreateTaskRequest{Description: "No title"})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("CreateTask() error code = %v; want %v", status.Code(err), codes.InvalidArgument)
		}
		if len(tm.service.tasks) != 0 {
			t.Errorf("CreateTask() saved a task for invalid input; task count = %d", len(tm.service.tasks))
		}
	})

	t.Run("assigns unique IDs", func(t *testing.T) {
		tm := newTaskManager()
		req := &pb.CreateTaskRequest{Title: "Task"}

		first, err := tm.CreateTask(context.Background(), req)
		if err != nil {
			t.Fatalf("first CreateTask() error = %v", err)
		}
		second, err := tm.CreateTask(context.Background(), req)
		if err != nil {
			t.Fatalf("second CreateTask() error = %v", err)
		}
		if first.GetId() == second.GetId() {
			t.Errorf("CreateTask() returned duplicate ID %q", first.GetId())
		}
	})
}
