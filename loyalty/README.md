Loyalty - сервис начисления и списания балов лояльности


```bash
curl --header "Content-Type: application/json" --request POST --data '{"uuid":"ea12bd7b-5d6d-4aa4-986c-64719186f742","value":100}' http://localhost:8001/loyalty/
```
```bash
cd commands
go run ./migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
```