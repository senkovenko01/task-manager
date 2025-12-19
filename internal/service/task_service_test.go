package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"task-manager/internal/models"
	"task-manager/internal/repository"
)

type inMemoryRepo struct {
	store map[uuid.UUID]*models.Task
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{store: make(map[uuid.UUID]*models.Task)}
}

func (r *inMemoryRepo) Ping(ctx context.Context) error {
	return nil
}

func (r *inMemoryRepo) CreateTask(ctx context.Context, task *models.Task) error {
	r.store[task.ID] = task
	return nil
}

func (r *inMemoryRepo) GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	if task, ok := r.store[taskID]; ok {
		return task, nil
	}
	return nil, repository.ErrTaskNotFound
}

func (r *inMemoryRepo) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*models.Task, error) {
	var tasks []*models.Task
	for _, task := range r.store {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *inMemoryRepo) UpdateTask(ctx context.Context, task *models.Task) error {
	if _, ok := r.store[task.ID]; !ok {
		return repository.ErrTaskNotFound
	}
	r.store[task.ID] = task
	return nil
}

func (r *inMemoryRepo) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	if _, ok := r.store[taskID]; !ok {
		return repository.ErrTaskNotFound
	}
	delete(r.store, taskID)
	return nil
}

func TestCreateTaskValidation(t *testing.T) {
	repository := newInMemoryRepo()
	service := NewTaskService(repository)

	_, err := service.CreateTask(context.Background(), models.CreateTaskInput{
		Title:       "ab",
		Description: "desc",
	})
	if err == nil {
		t.Fatal("expected validation error for short title")
	}
}

func TestUpdateTaskStatusValidation(t *testing.T) {
	repository := newInMemoryRepo()
	service := NewTaskService(repository)

	createdTask, err := service.CreateTask(context.Background(), models.CreateTaskInput{
		Title:       "valid title",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	badStatus := models.TaskStatus("wrong")
	_, err = service.UpdateTask(context.Background(), createdTask.ID, models.UpdateTaskInput{
		Status: &badStatus,
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}
