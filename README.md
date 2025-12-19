## Task Manager

Simple task management REST API.

### Why SQLite

- **Local-first**: Easy to run locally without external services.
- **Single binary**: DB is just a file (`tasks.db`) next to the app, good for a small test project.
- **Migrations**: Schema is created on startup via migrations package.

PostgreSQL would be a better choice for production, but SQLite keeps this repo easy to clone and run.


### Run Locally

```bash
# Build and start with docker-compose
docker-compose up --build
```

The server will start on `:8080` by default.

### Environment Variables

- `TASK_MANAGER_ADDR` – HTTP listen address (default `:8080`)
- `TASK_MANAGER_SQLITE_PATH` – SQLite DB file path (default `tasks.db`)
- `SEED_DATA` – Set to `true` to populate database with 25 sample tasks on startup (default `false`)

### Seed Data

The application includes a seed function that creates 25 sample tasks with various statuses (`new`, `in_progress`, `done`). To enable seeding:

### API

All endpoints use JSON.

- **Create task**

  - `POST /tasks`
  - Body:

    ```json
    {
      "title": "Buy milk",
      "description": "2 liters"
    }
    ```

  - Rules:
    - `title` required, minimum 3 characters
    - `status` defaults to `new`

- **List tasks**

  - `GET /tasks`
  - Query params:
    - `status` (optional) – `new | in_progress | done`
    - `limit` (optional, default 50)
    - `offset` (optional, default 0)

- **Get task by ID**

  - `GET /tasks/{id}`

- **Update task**

  - `PUT /tasks/{id}`
  - Body:

    ```json
    {
      "title": "Buy milk and bread",
      "description": "2 liters + baguette",
      "status": "in_progress"
    }
    ```

  - Any field can be omitted; provided fields are validated.

- **Delete task**

  - `DELETE /tasks/{id}`

- **Health check**

  - `GET /health`

### Architecture

- **`main.go`**
  - App entrypoint
  - Loads config, opens SQLite DB, runs migrations
  - Optionally seeds database with sample data if `SEED_DATA=true`
  - Builds repository, service, and HTTP handlers
  - Starts HTTP server with graceful shutdown

- **`internal/migrations`**
  - `migrations.go` – Database schema migrations
  - `seed.go` – Seed data function (creates 25 sample tasks)

- **`internal/models`**
  - Domain models (`Task`, `TaskStatus`)
  - Input DTOs (`CreateTaskInput`, `UpdateTaskInput`)

- **`internal/repository`**
  - `TaskRepository` interface
  - `SQLiteTaskRepository` implementation using `database/sql`
  - All methods accept `context.Context` and map `sql.ErrNoRows` to domain errors

- **`internal/service`**
  - Business logic:
    - Validation for title length and status values
    - Default values on create
  - Works only with `TaskRepository` interface (no HTTP or SQL details)

- **`internal/handler`**
  - HTTP transport (REST)
  - JSON decoding/encoding
  - Query parameter parsing (status, limit, offset)
  - Maps domain/service errors to HTTP status codes

- **`internal/config`**
  - Simple env-based configuration loader

### Notes on decisions

- **Context**
  - Every DB call uses `context.Context` (through `database/sql` context-aware methods).
  - HTTP handlers pass `r.Context()` into the service layer so request cancellation propagates down to DB.

- **Error handling**
  - Repository exposes a typed `ErrTaskNotFound` error for 404 mapping.
  - Validation errors are surfaced as `400 Bad Request` with human-readable messages.
  - No `panic` in business logic – only in startup failures where the app cannot continue.

- **Pagination**
  - Implemented via `limit` and `offset` query params on `GET /tasks`.


### Manual Checks that can be performed

```bash
# Health check
curl http://localhost:8080/health

# Create a new task
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy milk", "description": "2 liters"}'

# List all tasks
curl "http://localhost:8080/tasks"

# List tasks with filters
curl "http://localhost:8080/tasks?status=new&limit=10&offset=0"

# Get a specific task (replace {id} with actual UUID)
curl "http://localhost:8080/tasks/{id}"

# Update a task
curl -X PUT http://localhost:8080/tasks/{id} \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy milk and bread", "status": "in_progress"}'

# Delete a task
curl -X DELETE http://localhost:8080/tasks/{id}
```


