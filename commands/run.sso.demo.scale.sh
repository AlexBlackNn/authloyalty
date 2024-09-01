#!/bin/bash
docker rm -f $(docker ps -aq)
cd ../infra && docker compose -f docker-compose.demo.scale.yaml up -d
timeout 90s bash -c "until docker exec patroni1 pg_isready ; do sleep 5 ; done"
cd ../commands && go run ./migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
cd ../infra && docker compose -f docker-compose.demo.scale.yaml scale sso=2