# SSO  - Сервис авторизации

Единый сервис авторизации (СА) и управления пользователями 

## Swagger grpc доступен по адресу:
http://127.0.0.1:44044/sw/

## Swagger http доступен по адресу:
http://127.0.0.1:8000/swagger/index.html

## Архитектурные решения 

### API:
HTTP-handlers используются для общения с frontend, а gRPC для общения внутри микросервисов.
* gRPC: gRPC использует Protobuf для сериализации данных, что обеспечивает более компактное
представление сообщения по сравнению с JSON (используемым в HTTP), следовательно
реализуется более быстрый обмен данными.
* HTTP: HTTP, является стандартом для веб-разработки, что делает его простым и легким для
интеграции с frontend-фреймворками.


## Технический анализ особенностей регистрации пользователей

### 1. Немедленная выдача токенов

При регистрации пользователя система сразу выдает access и refresh токены, минуя этап входа в систему. Это позволяет пользователю начать взаимодействие с сервисом непосредственно после регистрации.

### 2. Использование Patroni кластера для хранения данных

Система использует Patroni кластер, обеспечивающий высокую доступность данных за счет репликации.

### 3. Асинхронная репликация и конечная согласованность

Репликация в Patroni кластере происходит асинхронно, что может привести к временной несогласованности между ведущим узлом и репликами. В результате, данные о новом пользователе могут отсутствовать на реплике в течение нескольких секунд после его регистрации.

### 4. Время жизни токенов и задержка репликации

Время жизни токенов установлено на 15 минут, что, как ожидается, будет достаточно для завершения репликации данных о пользователе на все реплики.

### 5. Временная несогласованность и ее влияние

Временная несогласованность данных на репликах, как правило, является кратковременным явлением. Она может привести к тому, что пользователь не сможет войти в систему сразу после регистрации, но это не должно быть проблемой, так как к моменту окончания срока действия токенов, репликация данных должна быть завершена.

### 6. Снижение нагрузки на БД за счет использования реплик

Чтение данных осуществляется с реплик, что значительно снижает нагрузку на ведущий узел. Это позволяет обеспечить более высокую производительность системы.

### 7. Возможные проблемы с задержкой репликации

Несмотря на то, что задержка репликации в штатном режиме невелика, при больших нагрузках или проблемах с сетью она может быть существенной. В этом случае пользователь может столкнуться с проблемами при авторизации. Вероятность такой ситуации мала, так как сильное отставание реплики в стабильно работающей системе встречается редко.

### 8. Компромисс между производительностью и согласованностью

Использование реплик для чтения данных представляет собой разумный компромисс между производительностью и согласованностью. Несмотря на риск временной несогласованности, этот подход позволяет обеспечить более высокую производительность системы и улучшить пользовательский опыт.





cd authloyalty/protos/proto/registration
protoc --go_out=. registration.proto

easyjson -all /home/alex/Dev/GolandYandex/authloyalty/internal/handlersapi/v1/sso_handlers_response.go


```bash
curl --header "Content-Type: application/json" --request POST --data '{"email":"test@test.com","password":"test"}' http://localhost:8000/auth/login
```

```bash
curl --header "Content-Type: application/json" --request POST --data '{"email":"test@test.com","password":"test"}' http://localhost:8000/auth/registration
```

```bash
curl --header "Content-Type: application/json" --request POST --data '{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjI0MzIwODMsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOjJ9.J6XilG2yEAM611yybY8LdvXs046yrx8bjCoWlwd5dtQ"}' http://localhost:8000/auth/logout
```

// HOW TO ADD GRPC SWAGGER
https://apidog.com/articles/how-to-add-swagger-ui-for-grpc/

1. download buf bin from github
2. rename to buf
3. move to /usr/bin
4. chmod +x buf
5. buf generate 

if Failure: plugin openapiv2: could not find protoc plugin for name openapiv2 - please make sure protoc-gen-openapiv2 is installed and present on your $PATH
```bash
go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

# redis sentinel
https://redis.uptrace.dev/guide/go-redis-sentinel.html#redis-server-client

# auth swagger
http://localhost:8000/swagger/index.html

```
swag init -g ./cmd/sso/main.go -o ./cmd/sso/docs
```

if err when starts 

Golang swaggo rendering error: "Failed to load API definition" and "Fetch error doc.json" [closed]
Where the routers locate
n most cases, the problem is that you forgot to import the generated docs as _ "<your-project-package>/docs" 
in my case
_ "github.com/AlexBlackNn/authloyalty/cmd/sso/docs"


metrics
https://stackoverflow.com/a/65609042
https://github.com/prometheus/client_golang
https://grafana.com/oss/prometheus/exporters/go-exporter/#metrics-usage
