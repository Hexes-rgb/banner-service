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
	docker volume ls -q | grep pgdata | xargs -r docker volume rm
