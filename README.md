# Сервис начисления лояльности с авторизацией 

### Запуск

#### Необходимое ПО:
1. [Task](https://taskfile.dev/installation/) -  sudo snap install task --classic (ubuntu 22.04)
2. [Docker](https://docs.docker.com/engine/install/)
3. [Protocol Buffer Compiler](https://grpc.io/docs/protoc-installation/#install-using-a-package-manager) - если нужно генерить protobuf


## Запуск тестов
1. Юнит тесты
```bash
cd commands && task unit-tests
```

2. Интеграционные тесты
```bash
cd commands && task integration-tests
```

##  Доступны два варианта запуска: локальный и на демо стенде.

##### Локальный запуск:
* Разворачивается инфраструктура в `docker compose`.
* Сервисы запускаются локально.

1. Настройка инфраструктуры
```bash
cd commands && task local
```

2. Запуск сервиса авторизации
```bash
go run ./sso/cmd/sso/main.go --config=./sso/config/local.yaml
```

3. Запуск сервиса начисления балов лояльности
```bash
go run ./loyalty/cmd/main.go --config=./loyalty/config/local.yaml
```

4. Взаимодействие между сервисами
    
    4.1. Открыть swagger   
    a. [swagger](http://localhost:8000/swagger/index.htm/index.html) сервиса авторизации
    
    б. [swagger](http://localhost:8001/swagger/index.htm/index.html) сервиса начисления балов лояльности 

    4.2 В swagger сервиса авторизации зарегестрироваться.
   ![registration.png](docs%2Fregistration.png)
    
    Ожидаемый вывод: 
    ```
    {
    "status": "Success",
    "user_id": "8c3201a2-8a04-4d4d-a382-74254f95acee",
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjUxNzM1MjMsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiI4YzMyMDFhMi04YTA0LTRkNGQtYTM4Mi03NDI1NGY5NWFjZWUifQ.Vs0_3XGKETR5roWuogS46YgxRo_rcW5KtYz5z_ACgG8",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjYwMzM5MjMsInRva2VuX3R5cGUiOiJyZWZyZXNoIiwidWlkIjoiOGMzMjAxYTItOGEwNC00ZDRkLWEzODItNzQyNTRmOTVhY2VlIn0.hu4JOd30HxyHWJBDUP2H0b1IDlVfNmGP0lPh42lghmk"
    }
    ```
   
    4.3 В swagger сервиса начисления былов лояльности нажать Authorize и ввести access_token из пункта 4.2
    