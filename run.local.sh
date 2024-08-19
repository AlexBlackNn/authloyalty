#!/bin/bash
docker rm -f $(docker ps -aq)
cd ../infra && docker-compose -f docker-compose.local.yaml up -d
cd .. && go run ./cmd/migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
go run ./cmd/sso/main.go --config=./config/local.yaml
go run ./kafka_consumer/main.go --config=./config/local.yaml