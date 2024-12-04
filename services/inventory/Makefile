# Makefile

include .env

postgresinit:
	@docker-compose up -d

createdb:
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"


dropdb:
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);"


migrateup:
	@migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" -verbose up


migratedown:
	@migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" -verbose down
