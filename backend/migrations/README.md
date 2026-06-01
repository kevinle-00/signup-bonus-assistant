# Migrations

These SQL files use Goose annotations.

Install Goose when you are ready to run migrations:

```sh
go install github.com/pressly/goose/v3/cmd/goose@v3.24.3
```

Example command once Postgres is running:

```sh
goose -dir backend/migrations postgres "$DATABASE_URL" up
```
