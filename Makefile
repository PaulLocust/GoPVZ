.env:
	@touch .env

include .env
export

.PHONY: all test build clean help

all: test build

# --- Testing Section ---
.PHONY: test test-verbose coverage

test:
	@go test \
		./internal/auth/usecase \
		./internal/auth/repo \
		./internal/auth/controller/http \
		./internal/pvz/usecase \
		./internal/pvz/repo \
		./internal/pvz/controller/http \
		-coverprofile=coverage.out

test-verbose:
	@go test -v \
		./internal/auth/usecase \
		./internal/auth/repo \
		./internal/auth/controller/http \
		./internal/pvz/usecase \
		./internal/pvz/repo \
		./internal/pvz/controller/http \
		-coverprofile=coverage.out

coverage:
	@echo "Generating combined coverage report..."
	@go tool cover -html=coverage.out

# --- Application Commands ---
.PHONY: run migrate-up migrate-down migrate-create

run: .env
	@echo "Trying to start application..."
	@go run cmd/GoPVZ/main.go

migrate-up: .env
	@echo "Applying database migrations..."
	@migrate -path=${MIGRATIONS_PATH} -database "${PG_URL}" -verbose up

migrate-down: .env
	@echo "Reverting database migrations..."
	@migrate -path=${MIGRATIONS_PATH} -database "${PG_URL}" -verbose down

migrate-create: .env
	@echo "Creating new migration file..."
	@migrate create -ext=sql -dir=${MIGRATIONS_PATH} -seq ${name}

# --- Code Generation ---
.PHONY: generate-dto generate-swagger

generate-dto:
	@echo "Generating DTO types from OpenAPI spec..."
	@oapi-codegen -generate types -package dto -o internal/dto/types.gen.go api/swagger.yaml

generate-swagger:
	@echo "Generating Swagger documentation..."
	@swag init --generalInfo internal/app/app.go --output ./docs --parseDependency --parseInternal

# --- Utility Commands ---
clean:
	@echo "Cleaning up..."
	@rm -f ./internal/auth/coverage* ./*.log
	@docker system prune -f

help:
	@echo "Available commands:"
	@echo ""
	@echo "Application:"
	@echo "  make run              - Start the application"
	@echo "  make migrate-up       - Apply database migrations"
	@echo "  make migrate-down     - Revert database migrations"
	@echo "  make migrate-create   - Create new migration file"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run all tests without verbose"
	@echo "  make test-verbose     - Run all tests with verbose"
	@echo "  make coverage         - Generate coverage report"
	@echo "  Note: Docker desktop must be running for tests"
	@echo ""
	@echo "Code Generation:"
	@echo "  make generate-dto     - Generate DTO types"
	@echo "  make generate-swagger - Generate Swagger docs"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean            - Clean temporary files"
	@echo "  make help             - Show this help"