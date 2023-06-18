init-db:
	docker exec -it ${DB_CONTAINER} createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}

db:
	docker exec -it ${DB_CONTAINER} psql

migration:
	@read -p "Enter migration name: " name; \
		migrate create -ext sql -dir db/migrations $$name

migrate:
	migrate -source file://db/migrations \
		-database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE} up

rollback:
	migrate -source file://db/migrations \
		-database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE} down

drop:
	migrate -source file://db/migrations \
		-database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE} drop

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

.PHONY: init-db migrate db drop migration rollback sqlc test