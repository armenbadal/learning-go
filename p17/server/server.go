package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	pb "etm/gen/etasks"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	EmptyTitle   = errors.New("empty title")
	EmptyID      = errors.New("empty ID")
	TaskNotFound = errors.New("task not found")
)

type Task struct {
	ID          string
	Title       string
	Description string
	Completed   bool
}

type TaskService struct {
	mx    sync.Mutex
	index int32
	tasks map[string]*Task
}

func (t *TaskService) Create(title, description string) (string, error) {
	t.mx.Lock()
	defer t.mx.Unlock()

	if title == "" {
		return "", EmptyTitle
	}

	t.index++
	task := &Task{
		ID:          fmt.Sprintf("TASK-%d", t.index),
		Title:       title,
		Description: description,
		Completed:   false,
	}
	t.tasks[task.ID] = task

	return task.ID, nil
}

func (t *TaskService) Get(id string) (*Task, error) {
	t.mx.Lock()
	defer t.mx.Unlock()

	if id == "" {
		return nil, EmptyID
	}

	task, ok := t.tasks[id]
	if !ok {
		return nil, TaskNotFound
	}

	return task, nil
}

func (t *TaskService) List() ([]*Task, error) {
	t.mx.Lock()
	defer t.mx.Unlock()

	var tasks []*Task
	for _, task := range t.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t *TaskService) Complete(id string) error {
	t.mx.Lock()
	defer t.mx.Unlock()

	if id == "" {
		return EmptyID
	}

	if task, ok := t.tasks[id]; ok {
		task.Completed = true
		return nil
	}

	return TaskNotFound
}

func (t *TaskService) Delete(id string) error {
	t.mx.Lock()
	defer t.mx.Unlock()

	if id == "" {
		return EmptyID
	}

	if _, exists := t.tasks[id]; exists {
		delete(t.tasks, id)
		return nil
	}

	return TaskNotFound
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type TaskManager struct {
	service    *TaskService
	clients    sync.Map
	taskEvents chan *pb.TaskEvent
	pb.UnimplementedTaskManagerServer
}

const subscriberQueueSize = 10

type taskEventSubscriber struct {
	mu      sync.Mutex
	events  chan *pb.TaskEvent
	done    chan struct{}
	stopped bool
}

func newTaskEventSubscriber() *taskEventSubscriber {
	return &taskEventSubscriber{
		events: make(chan *pb.TaskEvent, subscriberQueueSize),
		done:   make(chan struct{}),
	}
}

func (s *taskEventSubscriber) enqueue(event *pb.TaskEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return false
	}

	select {
	case s.events <- event:
		return true
	default:
		s.stopped = true
		close(s.done)
		return false
	}
}

func (tm *TaskManager) CreateTask(ctx context.Context, rq *pb.CreateTaskRequest) (*pb.TaskIDResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if rq == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	id, err := tm.service.Create(rq.GetTitle(), rq.GetDescription())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tm.taskEvents <- &pb.TaskEvent{
		Type:      pb.TaskEventType_TASK_EVENT_TYPE_CREATED,
		TaskId:    id,
		Timestamp: timestamppb.Now(),
	}

	return &pb.TaskIDResponse{Id: id}, nil
}

func (tm *TaskManager) GetTask(ctx context.Context, rq *pb.TaskIDRequest) (*pb.TaskResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if rq == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	task, err := tm.service.Get(rq.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.TaskResponse{
		Id:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
	}, nil
}

func (tm *TaskManager) ListTasks(ctx context.Context, in *emptypb.Empty) (*pb.TaskListResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	tasks, err := tm.service.List()
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	result := &pb.TaskListResponse{}
	for _, task := range tasks {
		r := &pb.TaskResponse{
			Id:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Completed:   task.Completed,
		}
		result.Tasks = append(result.Tasks, r)
	}
	return result, nil
}

func (tm *TaskManager) CompleteTask(ctx context.Context, in *pb.TaskIDRequest) (*emptypb.Empty, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	err := tm.service.Complete(in.GetId())
	if err != nil {
		if err == TaskNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tm.taskEvents <- &pb.TaskEvent{
		Type:      pb.TaskEventType_TASK_EVENT_TYPE_COMPLETED,
		TaskId:    in.GetId(),
		Timestamp: timestamppb.Now(),
	}

	return &emptypb.Empty{}, nil
}

func (tm *TaskManager) DeleteTask(ctx context.Context, in *pb.TaskIDRequest) (*emptypb.Empty, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	err := tm.service.Delete(in.GetId())
	if err != nil {
		if err == TaskNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tm.taskEvents <- &pb.TaskEvent{
		Type:      pb.TaskEventType_TASK_EVENT_TYPE_DELETED,
		TaskId:    in.GetId(),
		Timestamp: timestamppb.Now(),
	}

	return &emptypb.Empty{}, nil
}

func (tm *TaskManager) WatchTaskEvents(req *pb.SubscribeRequest, stream pb.TaskManager_WatchTaskEventsServer) error {
	clientName := req.GetName()
	log.Printf("Client '%s' requested connection", clientName)

	subscriber := newTaskEventSubscriber()
	tm.clients.Store(clientName, subscriber)

	defer func() {
		tm.clients.CompareAndDelete(clientName, subscriber)
	}()

	for {
		select {
		case event := <-subscriber.events:
			if err := stream.Send(event); err != nil {
				log.Printf("failed to send task event to client %q: %v", clientName, err)
				return err
			}
		case <-subscriber.done:
			return status.Error(codes.ResourceExhausted, "subscriber event queue is full")
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	tm := TaskManager{
		service: &TaskService{
			index: 0,
			tasks: make(map[string]*Task),
		},
		taskEvents: make(chan *pb.TaskEvent, 10),
	}
	pb.RegisterTaskManagerServer(server, &tm)
	log.Printf("Server listening at %v", listener.Addr())

	go func() {
		for event := range tm.taskEvents {
			tm.clients.Range(func(clientName, client any) bool {
				subscriber, ok := client.(*taskEventSubscriber)
				if !ok {
					tm.clients.Delete(clientName)
					return true
				}

				if !subscriber.enqueue(event) {
					log.Printf("disconnecting slow client %q: event queue is full", clientName)
					tm.clients.CompareAndDelete(clientName, subscriber)
				}
				return true
			})
		}
	}()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
