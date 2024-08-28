Loyalty - сервис начисления и списания балов лояльности

//addloyalty
```bash
curl --header "Content-Type: application/json" --request POST --data '{"uuid":"39d31ff1-e1f0-4eda-b52c-a68e2ab3eae7","value":10}' http://localhost:8001/loyalty/
```

//getloyalty
```bash
curl http://localhost:8001/loyalty/07cf0415-a5ca-4b2e-ba73-761ff31173fd
```

```bash
cd commands
go run ./migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
```