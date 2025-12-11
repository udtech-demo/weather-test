hello:
	fiber "Hello"

start: dependents
	go mod download
	go run cmd/api/main.go

dependents:
	docker compose up -d
	brew install golang-migrate
	migrate -path src/infrastructure/storage/postgres/migrations -database postgres://admin:12345@localhost:5454/gbc?sslmode=disable up

run: dependents
	go run cmd/api/main.go
