run:
	nodemon --signal SIGTERM

start-postgres:
	docker run -d --rm --name postgres -e POSTGRES_USER=root -p 5432:5432 -e POSTGRES_PASSWORD=secret  -e POSTGRES_DB=caroption_go -e PGDATA=/var/lib/postgresql/data/pgdata -v "caroption_go":/var/lib/postgresql/data postgres:17.5

stop-postgres:
	docker stop postgres

migrate-up:
	migrate -path ./internal/db/migrations -database postgres://root:secret@localhost:5432/caroption_go?sslmode=disable up

migrate-down:
	migrate -path ./internal/db/migrations -database postgres://root:secret@localhost:5432/caroption_go?sslmode=disable down

migrate-fix:
	migrate -path ./internal/db/migrations -database postgres://root:secret@localhost:5432/caroption_go?sslmode=disable force 1


start-pgadmin:
	docker run -d --rm \
  --name pgadmin4 \
  -e PGADMIN_DEFAULT_EMAIL=${PGADMIN_DEFAULT_EMAIL} \
  -e PGADMIN_DEFAULT_PASSWORD=${PGADMIN_DEFAULT_PASSWORD} \
  -e SCRIPT_NAME=/pgadmin \
  -p 5050:80 \
  dpage/pgadmin4

stop-pgadmin:
	docker stop pgadmin4

test:
	go test **/**
