package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"task-manager/internal/models"
	"task-manager/internal/repository"
)

const DefaultLimit = 50

type TaskService interface {
	CreateTask(ctx context.Context, input models.CreateTaskInput) (*models.Task, error)
	GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	ListTasks(ctx context.Context, status *models.TaskStatus, limit, offset int) ([]*models.Task, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, input models.UpdateTaskInput) (*models.Task, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	Ping(ctx context.Context) error
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *taskService) CreateTask(ctx context.Context, input models.CreateTaskInput) (*models.Task, error) {
	if len(input.Title) < 3 {
		return nil, fmt.Errorf("title must be at least 3 characters")
	}
	task := &models.Task{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      models.TaskStatusNew,
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *taskService) ListTasks(ctx context.Context, status *models.TaskStatus, limit, offset int) ([]*models.Task, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	filter := repository.TaskFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	return s.repo.ListTasks(ctx, filter)
}

func (s *taskService) UpdateTask(ctx context.Context, taskID uuid.UUID, input models.UpdateTaskInput) (*models.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		if len(*input.Title) < 3 {
			return nil, fmt.Errorf("title must be at least 3 characters")
		}
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Status != nil {
		switch *input.Status {
		case models.TaskStatusNew, models.TaskStatusInProgress, models.TaskStatusDone:
			task.Status = *input.Status
		default:
			return nil, fmt.Errorf("invalid status")
		}
	}
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	return s.repo.DeleteTask(ctx, taskID)
}
