package migrations

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"task-manager/internal/models"
)

// SeedTasks inserts sample tasks into the database
func SeedTasks(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tasks := []struct {
		title       string
		description string
		status      models.TaskStatus
	}{
		{"Buy groceries", "Milk, eggs, bread, and vegetables", models.TaskStatusNew},
		{"Complete project report", "Finish the quarterly project report and submit to manager", models.TaskStatusInProgress},
		{"Call dentist", "Schedule annual checkup appointment", models.TaskStatusNew},
		{"Review code changes", "Review pull request #123 for the new feature", models.TaskStatusInProgress},
		{"Update documentation", "Update API documentation with latest endpoints", models.TaskStatusNew},
		{"Fix bug in login", "Investigate and fix authentication issue reported by users", models.TaskStatusInProgress},
		{"Plan team meeting", "Organize agenda and book conference room for next week", models.TaskStatusNew},
		{"Deploy to staging", "Deploy latest version to staging environment and run smoke tests", models.TaskStatusDone},
		{"Write unit tests", "Add unit tests for the new service layer", models.TaskStatusInProgress},
		{"Update dependencies", "Update Go modules and check for security vulnerabilities", models.TaskStatusNew},
		{"Design new feature", "Create mockups and technical design for user dashboard", models.TaskStatusNew},
		{"Optimize database queries", "Review and optimize slow queries in task repository", models.TaskStatusInProgress},
		{"Setup CI/CD pipeline", "Configure GitHub Actions for automated testing and deployment", models.TaskStatusDone},
		{"Refactor legacy code", "Refactor old authentication module to use new patterns", models.TaskStatusNew},
		{"Write blog post", "Draft blog post about Go best practices for the company blog", models.TaskStatusNew},
		{"Conduct code review", "Review and provide feedback on 5 pending pull requests", models.TaskStatusInProgress},
		{"Setup monitoring", "Configure Prometheus and Grafana for application metrics", models.TaskStatusDone},
		{"Create API documentation", "Generate OpenAPI spec and publish to documentation site", models.TaskStatusInProgress},
		{"Implement caching", "Add Redis caching layer for frequently accessed data", models.TaskStatusNew},
		{"Security audit", "Perform security audit and fix identified vulnerabilities", models.TaskStatusNew},
		{"Performance testing", "Run load tests and identify bottlenecks", models.TaskStatusInProgress},
		{"Update README", "Update project README with latest setup instructions", models.TaskStatusDone},
		{"Setup error tracking", "Integrate Sentry for error tracking and monitoring", models.TaskStatusNew},
		{"Create user guide", "Write comprehensive user guide for the application", models.TaskStatusNew},
		{"Backup database", "Create automated backup strategy for production database", models.TaskStatusDone},
	}

	now := time.Now().UTC()
	const checkQuery = `SELECT COUNT(*) FROM tasks WHERE title = ?`
	const insertQuery = `
INSERT INTO tasks (id, title, description, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
`

	for i, task := range tasks {
		// Check if a task with this title already exists
		var count int
		err := db.QueryRowContext(ctx, checkQuery, task.title).Scan(&count)
		if err != nil {
			return err
		}
		// Skip if task already exists
		if count > 0 {
			continue
		}

		// Spread tasks over the past few days for more realistic timestamps
		createdAt := now.Add(-time.Duration(i*3) * time.Hour)
		updatedAt := createdAt
		if task.status == models.TaskStatusInProgress || task.status == models.TaskStatusDone {
			updatedAt = createdAt.Add(time.Duration(i*2) * time.Hour)
		}

		id := uuid.New()
		_, err = db.ExecContext(ctx, insertQuery,
			id.String(),
			task.title,
			task.description,
			string(task.status),
			createdAt.Format(time.RFC3339Nano),
			updatedAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
