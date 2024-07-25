#!/bin/bash
docker rm -f $(docker ps -aq)
cd infra && docker-compose -f docker-compose.demo.yaml up -d
sleep 15
cd ..
go run ./cmd/migrator/postgres  --migrations-path=./migrations

