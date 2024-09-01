#!/bin/bash
docker rm -f $(docker ps -aq)
cd ../infra && docker compose -f docker-compose.demo.yaml up -d
timeout 90s bash -c "until docker exec patroni1 pg_isready ; do sleep 5 ; done"
pwd
cd ../commands && go run ./migrator/postgres  --p ./migrations_test -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
cd ../sso/tests/integragtion_tests && go test *.go -v