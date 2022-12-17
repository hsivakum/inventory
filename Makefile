build:
	docker compose -f docker-compose.yaml build --no-cache

run: build
	docker-compose -f docker-compose.yaml up -d

stop:
	docker-compose -f docker-compose.yaml down -v