Loyalty - сервис начисления и списания балов лояльности

//addloyalty
```bash
curl --header "Content-Type: application/json" --request POST --data '{"uuid":"ea12bd7b-5d6d-4aa4-986c-64719186f742","value":1}' http://localhost:8001/loyalty/
```

//getloyalty
```bash
curl http://localhost:8001/loyalty/e9d31ff1-e0f0-4eda-b52c-a68e2ab3eae7
```

```bash
cd commands
go run ./migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
```