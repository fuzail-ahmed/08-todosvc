package todo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("task not found")

// Repository defines data access operations for tasks.
type Repository interface {
	Create(ctx context.Context, t *Task) error
	GetByID(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context, page, pageSize int) ([]Task, int64, error) // returns items, total
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id string) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, t *Task) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}

func (r *gormRepository) GetByID(ctx context.Context, id string) (*Task, error) {
	var t Task
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by id: %w", err)
	}
	return &t, nil
}

func (r *gormRepository) List(ctx context.Context, page, pageSize int) ([]Task, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var tasks []Task
	var total int64
	q := r.db.WithContext(ctx).Model(&Task{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}
	offset := (page - 1) * pageSize
	if err := q.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, total, nil
}

func (r *gormRepository) Update(ctx context.Context, t *Task) error {
	// Save updates UpdatedAt automatically
	if err := r.db.WithContext(ctx).Model(&Task{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"title":       t.Title,
		"description": t.Description,
		"completed":   t.Completed,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (r *gormRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Task{}).Error; err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}
