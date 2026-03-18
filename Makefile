.PHONY: db-up db-down db-migrate db-reset backend-run

db-up: ## Start the database container (detached)
	docker compose up -d db

db-down: ## Stop all containers
	docker compose down

db-migrate: ## Run Flyway migrations
	docker compose run --rm flyway

db-reset: ## Tear down volumes, restart db, and re-run migrations
	docker compose down -v
	docker compose up -d db
	@echo "Waiting for db to be healthy..."
	@until docker compose exec db pg_isready -U $${POSTGRES_USER:-pf2e} -d $${POSTGRES_DB:-pf2e_companion}; do sleep 1; done
	docker compose run --rm flyway

backend-run: ## Run the backend API server locally
	set -a && . ./.env && set +a && cd backend && go run main.go
