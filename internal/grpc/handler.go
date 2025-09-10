package grpc

import (
	"context"

	"github.com/fuzail/08-todosvc/internal/todo"
	pb "github.com/fuzail/08-todosvc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type handler struct {
	pb.UnimplementedTodoServiceServer
	svc todo.Service
}

func NewHandler(s todo.Service) pb.TodoServiceServer {
	return &handler{svc: s}
}

func toProtoTask(t *todo.Task) *pb.Task {
	if t == nil {
		return nil
	}
	return &pb.Task{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func (h *handler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	t, err := h.svc.CreateTask(ctx, req.Title, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create: %v", err)
	}
	return &pb.CreateTaskResponse{Task: toProtoTask(t)}, nil
}

func (h *handler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	t, err := h.svc.GetTask(ctx, req.Id)
	if err != nil {
		if err == todo.ErrNotFound {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		return nil, status.Errorf(codes.Internal, "get: %v", err)
	}
	return &pb.GetTaskResponse{Task: toProtoTask(t)}, nil
}

func (h *handler) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	page := int(req.Page)
	if page == 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}
	tasks, total, err := h.svc.ListTasks(ctx, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list: %v", err)
	}
	protoTasks := make([]*pb.Task, 0, len(tasks))
	for _, t := range tasks {
		// copy to avoid pointer pitfalls
		copyT := t
		protoTasks = append(protoTasks, toProtoTask(&copyT))
	}
	return &pb.ListTasksResponse{
		Tasks:    protoTasks,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Total:    total,
	}, nil
}

func (h *handler) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	t, err := h.svc.UpdateTask(ctx, req.Id, req.Title, req.Description)
	if err != nil {
		if err == todo.ErrNotFound {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		return nil, status.Errorf(codes.Internal, "update: %v", err)
	}
	return &pb.UpdateTaskResponse{Task: toProtoTask(t)}, nil
}

func (h *handler) MarkComplete(ctx context.Context, req *pb.MarkCompleteRequest) (*pb.MarkCompleteResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	t, err := h.svc.MarkComplete(ctx, req.Id, req.Completed)
	if err != nil {
		if err == todo.ErrNotFound {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		return nil, status.Errorf(codes.Internal, "mark complete: %v", err)
	}
	return &pb.MarkCompleteResponse{Task: toProtoTask(t)}, nil
}

func (h *handler) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	if err := h.svc.DeleteTask(ctx, req.Id); err != nil {
		if err == todo.ErrNotFound {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		return nil, status.Errorf(codes.Internal, "delete: %v", err)
	}
	return &pb.DeleteTaskResponse{Success: true}, nil
}
