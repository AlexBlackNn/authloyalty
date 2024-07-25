#!/bin/bash
go run ../cmd/migrator/postgres  --migrations-path=./migrations
sleep 5
go run ../cmd/sso/main.go --config=../config/local.yaml &
sleep 5
go test auth_login_test.go
go test auth_register_login_test.go
go test auth_is_admin_test.go
go test auth_register_test.go
