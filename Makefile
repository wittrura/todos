compose: ## Start up API and database via docker compose
	docker compose up --build

compose-dev:
	docker compose \
	-f compose.yml \
	-f compose.dev.yml \
	-p todos-dev \
	up --build

compose-prod:
	docker compose \
	-f compose.yml \
	-f compose.prod.yml \
	-p todos-prod \
	up --build