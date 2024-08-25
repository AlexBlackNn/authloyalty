#!/bin/bash
docker rm -f $(docker ps -aq)
cd ../infra && docker compose -f docker-compose.prod.scale.yaml up -d
timeout 90s bash -c "until docker exec patroni1 pg_isready ; do sleep 5 ; done"
cd .. && go run ./sso/cmd/migrator/postgres  --p ./sso/migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
cd ./infra && docker compose -f docker-compose.prod.scale.yaml scale sso=2