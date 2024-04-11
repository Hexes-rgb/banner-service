build:
	@echo "Building services..."
	docker compose build

run:
	@echo "Running services..."
	docker compose up -d

stop:
	@echo "Stopping services..."
	docker compose stop
