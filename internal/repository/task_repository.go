package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-manager/internal/models"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskFilter struct {
	Status *models.TaskStatus
	Limit  int
	Offset int
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	Ping(ctx context.Context) error
}

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{db: db}
}

func (r *SQLiteTaskRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *SQLiteTaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	const query = `
INSERT INTO tasks (id, title, description, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
`
	_, err := r.db.ExecContext(ctx, query,
		task.ID.String(),
		task.Title,
		task.Description,
		string(task.Status),
		task.CreatedAt.Format(time.RFC3339Nano),
		task.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *SQLiteTaskRepository) GetTask(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	const query = `
SELECT id, title, description, status, created_at, updated_at
FROM tasks
WHERE id = ?
`
	row := r.db.QueryRowContext(ctx, query, taskID.String())
	var task models.Task
	var statusStr string
	var createdAtStr, updatedAtStr string
	if err := row.Scan(&task.ID, &task.Title, &task.Description, &statusStr, &createdAtStr, &updatedAtStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	task.Status = models.TaskStatus(statusStr)

	var err error
	task.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	task.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse_updated_at: %w", err)
	}
	return &task, nil
}

func (r *SQLiteTaskRepository) ListTasks(ctx context.Context, filter TaskFilter) ([]*models.Task, error) {
	baseQuery := `
SELECT id, title, description, status, created_at, updated_at
FROM tasks
`
	queryArgs := []any{}
	if filter.Status != nil {
		baseQuery += "WHERE status = ? "
		queryArgs = append(queryArgs, string(*filter.Status))
	}
	baseQuery += "ORDER BY created_at DESC "
	if filter.Limit > 0 {
		baseQuery += "LIMIT ? "
		queryArgs = append(queryArgs, filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += "OFFSET ?"
		queryArgs = append(queryArgs, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		var statusStr string
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &statusStr, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		task.Status = models.TaskStatus(statusStr)
		task.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		task.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse_updated_at: %w", err)
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now().UTC()
	const query = `
UPDATE tasks
SET title = ?, description = ?, status = ?, updated_at = ?
WHERE id = ?
`
	result, err := r.db.ExecContext(ctx, query,
		task.Title,
		task.Description,
		string(task.Status),
		task.UpdatedAt.Format(time.RFC3339Nano),
		task.ID.String(),
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *SQLiteTaskRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	const query = `DELETE FROM tasks WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, taskID.String())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}
