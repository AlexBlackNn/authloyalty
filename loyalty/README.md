Loyalty - сервис начисления и списания балов лояльности

//addloyalty withdraw
```bash
curl --header "Content-Type: application/json" --request POST --data '{"uuid":"f6c11262-8660-436a-9fcb-f95554cfe51a","balance":10,"operation":"w","comment":"withdraw loyalty"}' http://localhost:8001/loyalty/
```

//addloyalty deposit
```bash
curl --header "Content-Type: application/json" --request POST --data '{"uuid":"f6c11262-8660-436a-9fcb-f95554cfe51a","balance":10,"operation":"d","comment":"add loyalty"}' http://localhost:8001/loyalty/
```


//getloyalty
```bash
curl http://localhost:8001/loyalty/f6c11262-8660-436a-9fcb-f95554cfe51a
```

```bash
cd commands
go run ./migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
```