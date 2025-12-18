export DATABASE_URL="postgres://postgres:@localhost:5432/postmatic_go_dev?sslmode=disable"

goose -dir migrations postgres "$DATABASE_URL" down