# tramplin

Go + Fiber API for the Tramplin platform.

## Run locally

```bash
go mod tidy
go run ./cmd/api
```

App starts on `http://localhost:8080`.

## Run with Docker

```bash
docker compose up --build
```

App is available at `http://localhost:8080`.

## Swagger

Swagger UI is available at:

- `http://localhost:8080/swagger/index.html`
- `http://localhost:8080/docs`

Swagger docs are generated from comments above handlers:

```bash
swag init -g cmd/api/main.go -o docs
```

## Database

Default PostgreSQL connection string:

```bash
postgres://tramplin:tramplin@localhost:5432/tramplin?sslmode=disable
```
