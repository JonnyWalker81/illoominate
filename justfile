# Illoominate - Task Runner
# Run `just` to see all available commands

set dotenv-load := true

# Default recipe - show help
default:
    @just --list

# ============================================
# Development
# ============================================

# Start all services (API + Web + Supabase)
dev:
    @echo "Starting all services..."
    just db-start &
    sleep 3
    just api &
    just web

# Start Go API server with hot reload
api:
    cd api && watchexec -r -e go -- go run ./cmd/server

# Start React frontend dev server
web:
    cd web && pnpm dev

# ============================================
# Database
# ============================================

# Start Supabase local development
db-start:
    supabase start

# Stop Supabase local development
db-stop:
    supabase stop

# Run database migrations
db-migrate:
    psql "$DATABASE_URL" -f api/internal/db/migrations/000001_initial_schema.up.sql

# Run migrations using supabase (alternative)
db-migrate-supabase:
    supabase db push

# Rollback last migration
db-rollback:
    psql "$DATABASE_URL" -f api/internal/db/migrations/000001_initial_schema.down.sql

# Reset database (WARNING: destroys all data)
db-reset:
    @echo "WARNING: This will destroy all data!"
    @read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
    supabase db reset

# Create a new migration file
db-new name:
    @mkdir -p api/internal/db/migrations
    @touch "api/internal/db/migrations/$(date +%Y%m%d%H%M%S)_{{name}}.up.sql"
    @touch "api/internal/db/migrations/$(date +%Y%m%d%H%M%S)_{{name}}.down.sql"
    @echo "Created migration files for {{name}}"

# Show migration status
db-status:
    psql "$DATABASE_URL" -c "SELECT version, dirty FROM schema_migrations;" 2>/dev/null || echo "No migrations applied yet"

# ============================================
# Code Generation
# ============================================

# Generate sqlc code from queries
gen:
    cd api && sqlc generate

# Generate all (sqlc + any other generators)
gen-all: gen
    @echo "All code generation complete"

# ============================================
# Landing Page (Cloudflare Pages + D1)
# ============================================

# Start landing page dev server
landing:
    cd landing && npm run dev

# Start landing with Cloudflare Workers (full stack)
landing-dev:
    cd landing && npm run build && wrangler pages dev ./dist --d1=DB=illoominate-waitlist

# Run landing page tests
test-landing:
    cd landing && npm test

# Run landing page tests with coverage
test-landing-cover:
    cd landing && npm run test:coverage

# Run landing page tests in watch mode
test-landing-watch:
    cd landing && npm run test:watch

# Build landing page
build-landing:
    cd landing && npm run build

# Lint landing page
lint-landing:
    cd landing && npm run lint

# Format landing page code
fmt-landing:
    cd landing && npm run format

# Initialize local D1 database
landing-db-init:
    cd landing && wrangler d1 execute illoominate-waitlist --local --file=./schema.sql

# Reset local D1 database
landing-db-reset:
    @echo "WARNING: This will destroy all local waitlist data!"
    @read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
    cd landing && rm -rf .wrangler/state/v3/d1
    cd landing && wrangler d1 execute illoominate-waitlist --local --file=./schema.sql

# Create D1 database (run once for production)
landing-db-create:
    cd landing && wrangler d1 create illoominate-waitlist

# Apply schema to production D1
landing-db-migrate:
    cd landing && wrangler d1 execute illoominate-waitlist --remote --file=./schema.sql

# Deploy landing page to Cloudflare Pages
landing-deploy:
    cd landing && npm run build && wrangler pages deploy ./dist

# Install landing page dependencies
install-landing:
    cd landing && npm install

# ============================================
# Testing
# ============================================

# Run all tests
test: test-api test-web test-landing

# Run Go API tests
test-api:
    cd api && go test -v -race -cover ./...

# Run Go API tests with coverage report
test-api-cover:
    cd api && go test -v -race -coverprofile=coverage.out ./...
    cd api && go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: api/coverage.html"

# Run React frontend tests
test-web:
    cd web && pnpm test

# Run tests in watch mode
test-watch:
    cd api && watchexec -e go -- go test -v ./...

# ============================================
# Linting & Formatting
# ============================================

# Run all linters
lint: lint-api lint-web lint-landing

# Lint Go code
lint-api:
    cd api && golangci-lint run ./...

# Lint React code
lint-web:
    cd web && pnpm lint

# Format all code
fmt: fmt-api fmt-web fmt-landing

# Format Go code
fmt-api:
    cd api && go fmt ./...
    cd api && goimports -w .

# Format React code
fmt-web:
    cd web && pnpm format

# ============================================
# Build
# ============================================

# Build all
build: build-api build-web build-landing

# Build Go API binary
build-api:
    cd api && go build -o bin/server ./cmd/server

# Build React frontend
build-web:
    cd web && pnpm build

# ============================================
# Docker
# ============================================

# Build API Docker image
docker-build-api:
    docker build -t illoominate-api ./api

# Build and push API Docker image
docker-push-api:
    docker build -t gcr.io/$GCP_PROJECT_ID/illoominate-api ./api
    docker push gcr.io/$GCP_PROJECT_ID/illoominate-api

# ============================================
# SDK Development
# ============================================

# Build web widget
sdk-web-build:
    cd sdks/web-widget && pnpm build

# Watch web widget for changes
sdk-web-dev:
    cd sdks/web-widget && pnpm dev

# ============================================
# Utilities
# ============================================

# Clean build artifacts
clean:
    rm -rf api/bin
    rm -rf api/coverage.out api/coverage.html
    rm -rf web/dist
    rm -rf landing/dist
    rm -rf sdks/web-widget/dist

# Install dependencies
install:
    cd api && go mod download
    cd web && pnpm install
    cd landing && npm install
    cd sdks/web-widget && pnpm install

# Update dependencies
update:
    cd api && go get -u ./...
    cd api && go mod tidy
    cd web && pnpm update
    cd sdks/web-widget && pnpm update

# Show environment info
info:
    @echo "Go version: $(go version)"
    @echo "Node version: $(node --version)"
    @echo "pnpm version: $(pnpm --version)"
    @echo "Database URL: $DATABASE_URL"
    @echo "Supabase URL: $SUPABASE_URL"

# Create a new project (for testing)
create-test-project:
    @echo "Creating test project..."
    curl -X POST http://localhost:8080/api/creator/projects \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TEST_JWT" \
        -d '{"name": "Test Project", "slug": "test-project"}'
