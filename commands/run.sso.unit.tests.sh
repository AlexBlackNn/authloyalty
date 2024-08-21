#!/bin/bash
docker rm -f $(docker ps -aq)
cd ../infra && docker compose -f docker-compose.local.yaml up -d
timeout 90s bash -c "until docker exec patroni1 pg_isready ; do sleep 5 ; done"
cd .. && go run ./cmd/migrator/postgres  --p ./tests/integragtion_tests/migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
pwd
cd ./tests/unit_tests && go test *.go -v