#!/bin/bash
docker rm -f $(docker ps -aq)
cd infra && docker-compose up -d
sleep 15
cd ..
go run ./cmd/migrator/postgres  --migrations-path=./migrations
#sleep 5
#go run ./cmd/sso/main.go --config=./config/local.yaml