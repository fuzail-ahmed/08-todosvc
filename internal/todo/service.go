package todo

import (
	"context"
)

type Service interface {
	CreateTask(ctx context.Context, title, description string) (*Task, error)
	GetTask(ctx context.Context, id string) (*Task, error)
	ListTasks(ctx context.Context, page, pageSize int) ([]Task, int64, error)
	UpdateTask(ctx context.Context, id, title, description string) (*Task, error)
	MarkComplete(ctx context.Context, id string, completed bool) (*Task, error)
	DeleteTask(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) CreateTask(ctx context.Context, title, description string) (*Task, error) {
	t := &Task{
		Title:       title,
		Description: description,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) GetTask(ctx context.Context, id string) (*Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListTasks(ctx context.Context, page, pageSize int) ([]Task, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

func (s *service) UpdateTask(ctx context.Context, id, title, description string) (*Task, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Title = title
	t.Description = description
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) MarkComplete(ctx context.Context, id string, completed bool) (*Task, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Completed = completed
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) DeleteTask(ctx context.Context, id string) error {
	// check existence to return ErrNotFound consistently
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
