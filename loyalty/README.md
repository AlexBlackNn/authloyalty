Loyalty - сервис начисления и списания балов лояльности

//addloyalty withdraw

```bash
curl --header "Content-Type: application/json" --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXJAdGVzdC5jb20iLCJleHAiOjE3MjUwMzQzMzIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiIzMmE5YjJjOC0yMGJmLTQyZGUtYTRmNi0xYzZiN2VkYTE1MDYifQ.2ZHcnGtPon77nVc6zwRL2jpy4AeoSXVt-TVkyrXBww8" --request POST --data '{"uuid":"32a9b2c8-20bf-42de-a4f6-1c6b7eda1506","balance":10,"operation":"w","comment":"withdraw loyalty"}' http://localhost:8001/loyalty/
```

```bash
curl --header "Content-Type: application/json" --header "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXJAdGVzdC5jb20iLCJleHAiOjE3MjUwMzQzMzIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiIzMmE5YjJjOC0yMGJmLTQyZGUtYTRmNi0xYzZiN2VkYTE1MDYifQ.2ZHcnGtPon77nVc6zwRL2jpy4AeoSXVt-TVkyrXBww8" --request POST --data '{"uuid":"32a9b2c8-20bf-42de-a4f6-1c6b7eda1506","balance":10,"operation":"d","comment":"deposit loyalty"}' http://localhost:8001/loyalty/
```

//addloyalty deposit
```bash
curl --header "Content-Type: application/json" --request POST --data '{"uuid":"f0111262-8660-436a-9fcb-f95554cfe51a","balance":10,"operation":"d","comment":"add loyalty"}' http://localhost:8001/loyalty/
```


//getloyalty
```bash
curl http://localhost:8001/loyalty/f6c11262-8660-436a-9fcb-f95554cfe51a
```

```bash
cd commands
go run ./migrator/postgres  --p ./migrations -d postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable
```



SELECT * FROM loyalty_app.accounts;
SELECT * FROM loyalty_app.loyalty_transactions WHERE account_uuid='f3c11262-8660-436a-9fcb-f95554cfe51a';

SELECT SUM(CASE WHEN transaction_type='d' THEN transaction_amount ELSE -transaction_amount END)
FROM loyalty_app.loyalty_transactions
WHERE account_uuid='f3c11262-8660-436a-9fcb-f95554cfe51a'; 