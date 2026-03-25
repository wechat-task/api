# WeChat Task API

Task management system API service built with Golang + Gin + GORM + PostgreSQL

## Local Development

```bash
# Start database and API
docker-compose up

# Run tests
go test ./...
```

## API Endpoints

- `GET /api/v1/tasks` - List all tasks
- `POST /api/v1/tasks` - Create a new task
- `GET /api/v1/tasks/:id` - Get a single task
- `PUT /api/v1/tasks/:id` - Update a task
- `DELETE /api/v1/tasks/:id` - Delete a task
- `PUT /api/v1/tasks/:id/complete` - Mark a task as completed

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
