.PHONY: setup db-up db-down db-migrate db-seed db-reset backend-run backend-build ui-dev ui-build start stop test test-server test-ui

setup: ## Install all project dependencies (Go modules + npm packages)
	@echo "→ Installing Go dependencies..."
	cd backend && go mod download
	@echo "✔ Go modules ready"
	@echo "→ Installing npm packages..."
	cd ui && npm install
	@echo "✔ npm packages ready"
	@echo ""
	@echo "══════════════════════════════════════"
	@echo "  Setup complete"
	@echo "══════════════════════════════════════"

db-up: ## Start the database container (detached)
	docker compose up -d db
	@echo "✔ Database running on port $${POSTGRES_PORT:-5432}"

db-down: ## Stop all containers
	docker compose down

db-migrate: ## Run Flyway migrations
	docker compose run --rm flyway

db-seed: ## Run seed data against a migrated database
	docker compose run --rm seed

db-reset: ## Tear down volumes, restart db, re-run migrations, and seed
	docker compose down -v
	docker compose up -d db
	@echo "Waiting for db to be healthy..."
	@until docker compose exec db pg_isready -U $${POSTGRES_USER:-pf2e} -d $${POSTGRES_DB:-pf2e_companion}; do sleep 1; done
	docker compose run --rm flyway
	docker compose run --rm seed

backend-run: ## Run the backend API server locally
	@echo "→ Starting backend on port $${PORT:-8080}..."
	set -a && . ./.env && set +a && cd backend && go run main.go

ui-dev: ## Start the frontend dev server
	@echo "→ Starting frontend dev server on port 5173..."
	cd ui && npm run dev

backend-build: ## Build the backend binary
	cd backend && go build -o bin/server .

ui-build: ## Build the frontend for production
	cd ui && npm run build

start: backend-build ui-build ## Start the full stack (db, migrations, backend, frontend)
	docker compose up -d db
	@echo "Waiting for database to be healthy..."
	@until docker compose exec db pg_isready -U $${POSTGRES_USER:-pf2e} -d $${POSTGRES_DB:-pf2e_companion} > /dev/null 2>&1; do sleep 1; done
	docker compose run --rm flyway
	docker compose run --rm seed
	set -a && . ./.env && set +a && cd backend && go run main.go &
	cd ui && npm run dev &
	@echo ""
	@echo "══════════════════════════════════════"
	@echo "  Stack running"
	@echo "──────────────────────────────────────"
	@echo "  Database → localhost:$${POSTGRES_PORT:-5432}"
	@echo "  Backend  → http://localhost:$${PORT:-8080}"
	@echo "  Frontend → http://localhost:5173"
	@echo "══════════════════════════════════════"
	@wait

test-server: ## Run backend tests with coverage (min 80%)
	@echo "→ Running backend tests..."
	cd backend && go test ./... -coverprofile=coverage.out -covermode=atomic \
		-coverpkg=./auth/...,./handlers/...,./middleware/...,./ot/...,./repositories/...,./services/... && \
	go tool cover -func=coverage.out && \
	awk '/^total:/ { gsub(/%/, "", $$3); if ($$3+0 < 80) { print "FAIL: coverage " $$3 "%% is below 80%% threshold"; exit 1 } else { print "PASS: coverage " $$3 "%%" } }' coverage.out

test-ui: ## Run frontend tests with coverage (min 80%)
	@echo "→ Running frontend tests..."
	cd ui && npx vitest run --coverage && \
	node -e " \
	  const fs = require('fs'); \
	  const report = fs.readFileSync('coverage/coverage-summary.json', 'utf8'); \
	  const summary = JSON.parse(report).total.statements.pct; \
	  console.log('Coverage: ' + summary + '%'); \
	  if (summary < 80) { console.error('FAIL: coverage below 80%'); process.exit(1); } \
	  else { console.log('PASS: coverage meets threshold'); } \
	"

test: test-server test-ui ## Run all tests

stop: ## Tear down the full stack (containers, volumes, build artifacts, running processes)
	@echo "Killing processes on ports 8080 and 5173..."
	-@lsof -ti :8080 | xargs kill -9 2>/dev/null || true
	-@lsof -ti :5173 | xargs kill -9 2>/dev/null || true
	docker compose down -v
	rm -rf backend/bin
	rm -rf ui/dist
	@echo ""
	@echo "══════════════════════════════════════"
	@echo "  Teardown complete"
	@echo "──────────────────────────────────────"
	@echo "  ✔ Processes on :8080 and :5173 killed"
	@echo "  ✔ Containers & volumes removed"
	@echo "  ✔ backend/bin cleaned"
	@echo "  ✔ ui/dist cleaned"
	@echo "══════════════════════════════════════"
