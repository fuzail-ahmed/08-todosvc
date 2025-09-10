package test

import (
	"context"
	"testing"

	"github.com/fuzail/08-todosvc/internal/todo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&todo.Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCreateAndGetTask(t *testing.T) {
	db := setupTestDB(t)
	repo := todo.NewGormRepository(db)
	svc := todo.NewService(repo)

	ctx := context.Background()
	created, err := svc.CreateTask(ctx, "test title", "desc")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Title != "test title" {
		t.Fatalf("unexpected title: %s", created.Title)
	}

	got, err := svc.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("id mismatch: %s vs %s", got.ID, created.ID)
	}
}
