build:
	@echo "Building services..."
	cp .env.example .env
	docker compose build

run:
	@echo "Running services..."
	docker compose up -d

stop:
	@echo "Stopping services..."
	docker compose stop

delete:
	@echo "Deleting services..."
	docker compose down
	docker volume rm banner_service_pgdata