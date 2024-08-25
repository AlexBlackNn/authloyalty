## SSO - Сервис авторизации

Важное замечание:
* Телеметрия Kafka продюсера основана на 
[opentelemetry-go-contrib](https://github.com/etf1/opentelemetry-go-contrib).
Модификации кода будут выложены в отдельный репозиторий (сейчас authloyalty/pkg/tracing/otelconfluent). Необходимость модификации вызвана
невозможностью сохранить трейс без разбиейний. Быстрый фикс. Код для обсуждения. Не для ревью 
* Тестовый консьюмер `kafka_consumer` будет удален в дальнейших версиях. Не для ревью.
* SSO - входит в общий проект по начислению балов лояльности пользователям при покупке и регистрации.
* НЕ закрывал порты на БД в docker-compose. В проде врятли кто будет использовать 
 docker-compose на 1 машине. Скорее всего это будет k8s или еще какой-то оркестратор,
 а stateful приложения, вероятно, будут вынесены из кубера (холивар).

![Схема.jpg](docs%2FScheme.jpg)

Единый сервис авторизации (СА) и управления пользователями

### Запуск

#### Необходимое ПО:
1. [Task](https://taskfile.dev/installation/) -  sudo snap install task --classic (ubuntu 22.04)
2. [Docker](https://docs.docker.com/engine/install/)
3. [Protocol Buffer Compiler](https://grpc.io/docs/protoc-installation/#install-using-a-package-manager) - если нужно генерить protobuf

#### Доступны два варианта запуска: локальный и на демо стенде.

##### Локальный запуск:
* Разворачивается инфраструктура в `docker compose`.
* Сервис запускается локально.
```bash
cd commands && task local
```

Ожидаемый вывод (может занять какое-то время)
```
.......                                                                                                                                                                      2.5s 
 ✔ Container infra-redis_sentinel1-1  Started                                                                                                                                                                         2.5s 
 ✔ Container infra-redis_sentinel2-1  Started                                                                                                                                                                         2.6s 
Error response from daemon: No such container: patroni1
Error response from daemon: No such container: patroni1
.....
.....
.....
Error response from daemon: No such container: patroni1
migrations applied successfully
time=2024-08-21T21:06:03.503+03:00 level=INFO source=/xxx/authloyalty/internal/middleware/gzipCompressor.go:52 msg="gzip compressor enabled" component=middleware/gzip
time=2024-08-21T21:06:03.503+03:00 level=INFO source=/xxx/authloyalty/internal/middleware/gzipDecompressor.go:44 msg="gzip decompressor middleware enabled" component=middleware/gzip
time=2024-08-21T21:06:03.503+03:00 level=INFO source=/xxx/authloyalty/internal/middleware/logger.go:16 msg="logger middleware enabled" component=middleware/logger
2024/08/21 21:06:03 INFO grpc server starting
2024-08-21T21:06:03.505+0300    INFO    boot/grpc_entry.go:971  Bootstrap grpcEntry     {"eventId": "f7d05c8d-ecc3-47e8-8897-2ed4490ce7d0", "entryName": "sso", "entryType": "gRPC"}
------------------------------------------------------------------------
endTime=2024-08-21T21:06:03.505428384+03:00
startTime=2024-08-21T21:06:03.505230242+03:00
elapsedNano=198142
timezone=MSK
ids={"eventId":"f7d05c8d-ecc3-47e8-8897-2ed4490ce7d0"}
app={"appName":"rk","appVersion":"","entryName":"sso","entryType":"gRPC"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"pc","localIP":"172.25.0.1","os":"linux","realm":"*","region":"*"}
payloads={"grpcPort":44044,"gwPort":44044,"swEnabled":true,"swPath":"/sw/"}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=Bootstrap
resCode=OK
eventStatus=Ended
EOE
2024/08/21 21:06:03 INFO http server starting
```

##### Запуск на демо стенде:
* Разворачивается инфраструктура и сервис в `docker compose`.
```bash
cd commands && task demo
```

Ожидаемый вывод (может занять какое-то время)
```
.......                                                                                                                                                                      2.5s 
 ✔ Container infra-redis_sentinel1-1  Started                                                                                                                                                                         2.5s 
 ✔ Container infra-redis_sentinel2-1  Started                                                                                                                                                                         2.6s 
Error response from daemon: No such container: patroni1
Error response from daemon: No such container: patroni1
.....
.....
.....
Error response from daemon: No such container: patroni1
migrations applied successfully
```

##### Запуск интеграционных тестов:
* Разворачивается инфраструктура и сервис в `docker compose`.
* Запускаются тесты.
```bash
cd commands && task integration-tests 
```

Ожидаемый вывод (может занять какое-то время)
```
 ✔ Container infra-redis_sentinel3-1  Started                                                                                                                                                                         2.8s 
 ✔ Container infra-redis_sentinel1-1  Started                                                                                                                                                                         2.7s 
 ✔ Container infra-sso-1              Started                                                                                                                                                                         3.2s 
Error response from daemon: No such container: patroni1
....
Error response from daemon: No such container: patroni1
migrations applied successfully
=== RUN   TestIsAdminHappyPath
=== PAUSE TestIsAdminHappyPath
=== RUN   TestLoginHappyPath
=== PAUSE TestLoginHappyPath
=== RUN   TestLoginFailCases
=== PAUSE TestLoginFailCases
=== RUN   TestRegisterLoginHappyPath
=== PAUSE TestRegisterLoginHappyPath
=== RUN   TestRegisterHappyPath
=== PAUSE TestRegisterHappyPath
=== RUN   TestDuplicatedRegistration
=== PAUSE TestDuplicatedRegistration
=== RUN   TestAuthRegisterFailCases
=== PAUSE TestAuthRegisterFailCases
=== RUN   TestSuite
time=2024-08-21T21:15:08.044+03:00 level=INFO source=/xxx/authloyalty/internal/middleware/gzipCompressor.go:52 msg="gzip compressor enabled" component=middleware/gzip
time=2024-08-21T21:15:08.044+03:00 level=INFO source=/xxx/authloyalty/internal/middleware/gzipDecompressor.go:44 msg="gzip decompressor middleware enabled" component=middleware/gzip
time=2024-08-21T21:15:08.044+03:00 level=INFO source=/home/xxx/authloyalty/internal/middleware/logger.go:16 msg="logger middleware enabled" component=middleware/logger
...
--- PASS: TestSuite (0.59s)
    --- PASS: TestSuite/TestHttpServerRegisterHappyPath (0.57s)
        --- PASS: TestSuite/TestHttpServerRegisterHappyPath/user_registration (0.56s)
=== CONT  TestIsAdminHappyPath
=== CONT  TestRegisterHappyPath
=== CONT  TestLoginFailCases
=== CONT  TestRegisterLoginHappyPath
=== CONT  TestDuplicatedRegistration
=== CONT  TestLoginHappyPath
=== CONT  TestAuthRegisterFailCases
=== RUN   TestLoginFailCases/Login_with_Empty_Password
=== RUN   TestAuthRegisterFailCases/Register_with_Empty_Password
=== RUN   TestLoginFailCases/Login_with_Empty_Email
=== RUN   TestAuthRegisterFailCases/Register_with_Empty_Email
=== RUN   TestLoginFailCases/Login_with_Both_Empty_Email_and_Password
=== RUN   TestAuthRegisterFailCases/Register_with_Both_Empty
--- PASS: TestLoginFailCases (0.00s)
    --- PASS: TestLoginFailCases/Login_with_Empty_Password (0.00s)
    --- PASS: TestLoginFailCases/Login_with_Empty_Email (0.00s)
    --- PASS: TestLoginFailCases/Login_with_Both_Empty_Email_and_Password (0.00s)
--- PASS: TestAuthRegisterFailCases (0.00s)
    --- PASS: TestAuthRegisterFailCases/Register_with_Empty_Password (0.00s)
    --- PASS: TestAuthRegisterFailCases/Register_with_Empty_Email (0.00s)
    --- PASS: TestAuthRegisterFailCases/Register_with_Both_Empty (0.00s)
--- PASS: TestIsAdminHappyPath (0.01s)
--- PASS: TestLoginHappyPath (0.06s)
--- PASS: TestRegisterHappyPath (0.07s)
--- PASS: TestDuplicatedRegistration (0.13s)
--- PASS: TestRegisterLoginHappyPath (0.13s)
PASS
ok      command-line-arguments  0.739s
```


##### Запуск юнит тестов:
* Инфраструктура НЕ разворачивается (сделаны моки).
* Запускаются тесты.
```bash
cd commands && task unit-tests 
```
Ожидаемый вывод 
```
time=2024-08-21T22:00:47.313+03:00 level=INFO source=/home/alex/Dev/GolandYandex/authloyalty/internal/middleware/logger.go:30 msg="request completed" component=middleware/logger method=POST path=/auth/registration remote_addr=127.0.0.1:44912 user_agent=Go-http-client/1.1 request_id=pc/0kfqBOKjvE-000001 status=201 bytes=365 duration=105.431685ms
--- PASS: TestSuite (0.16s)
    --- PASS: TestSuite/TestHttpServerRegisterHappyPath (0.11s)
        --- PASS: TestSuite/TestHttpServerRegisterHappyPath/user_registration (0.11s)
PASS
ok      command-line-arguments  0.178s

```

### Доступ к UI:

* [Swagger grpc](http://localhost:44044/sw/)
* [Swagger http](http://localhost:8000/swagger/index.html)

Пример:

![swagger_http.png](docs%2Fswagger_http.png)
* [Grafana](http://localhost:3000/grafana/) 

Для просмотра метрик. В папке `monitoring` лежит dashboard `6671_rev2.json`.

![metrics.png](docs%2Fmetrics.png)

Логирование. 

![logs.png](docs%2Flogs.png)

* [Prometheus](http://localhost:9090/targets/)
* [Jaeger](http://localhost:16686/jaeger/search)

Запустить косьмер в другом терминале, чтобы увидеть распростронение трейса между микросервисами
```bash
cd commands && task consumer
```
![tracing.png](docs%2Ftracing.png)

* [Kafka UI](http://localhost:8080)

![kafkaui.png](docs%2Fkafkaui.png)

## Использование Postman 

Регистрация 
![registration_grpc.png](docs%2Fregistration_grpc.png)

Логин
![login_grpc.png](docs%2Flogin_grpc.png)
----------------------------------------------------

## Архитектурные решения 

### API:
HTTP-handlers используются для общения с frontend, а gRPC для общения внутри микросервисов.
* gRPC: использует Protobuf для сериализации данных, что обеспечивает более компактное
представление сообщения по сравнению с JSON (используемым в HTTP), следовательно
реализуется более быстрый обмен данными.
* HTTP: стандарт для веб-разработки, что делает его простым и легким для
интеграции с frontend-фреймворками.


### Регистрация пользователей

Система реализует немедленную выдачу токенов при регистрации(толкьо в http API, в случае grpc handler 
регистрации будет удален), позволяя пользователю начать работу с сервисом сразу. Для обеспечения высокой доступности данных используется Patroni кластер с асинхронной репликацией, что может приводить к кратковременной несогласованности данных между ведущим узлом и репликами. Однако, время жизни токенов (15 минут) достаточно для завершения репликации. Чтение данных с реплик снижает нагрузку на ведущий узел, повышая производительность системы.

Несмотря на риск временной несогласованности, этот подход обеспечивает баланс между производительностью и согласованностью данных. Пользователь может столкнуться с кратковременными проблемами авторизации в случае больших нагрузок или сетевых проблем, но вероятность этого мала.

## Асинхронная обработка событий регистрации

Сервис использует Kafka для асинхронной обработки событий регистрации пользователей.
При успешной регистрации генерируется сообщение в Kafka-топик, содержащее информацию о новом пользователе. Сервис отслеживает статус доставки сообщений и обновляет его в базе данных.
Это позволяет другим сервисам, например, сервису отправки приветственных сообщений или начисления баллов лояльности, подписываться на этот топик и асинхронно обрабатывать информацию о новых пользователях.

## Выбор библиотеки для взаимодействия сервисом через Kafka

Для взаимодействия между сервисами через Kafka используется формат данных protobuf. Преимуществом по сравнению с 
json является снижение кол-ва передаваемых байт (см. Высоконагруженные приложения. Программирование, масштабирование, поддержка. Автор
Мартин Клеппман). 

Было принято решение передачи данных через брокер Kafka с использованием   Schema Registry.  Schema Registry - это сервис, который позволяет хранить 
и управлять схемами данных, которые используются в Kafka. Это позволит:
* Управлять версиями схем: можно добавлять новые поля, изменять типы данных и т.д. без нарушения совместимости.
* Проверять данные: Schema Registry гарантирует, что данные, передаваемые в Kafka, соответствуют определенной схеме.
* Упростить десериализацию: Получатели данных могут использовать Schema Registry для получения необходимой схемы для десериализации данных.

Рассматривались две библиотеки в go. [Kafka-go](https://github.com/segmentio/kafka-go) и [Сonfluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go).
Kafka-go данный момент не имеет встроенной поддержки Schema Registry https://github.com/segmentio/kafka-go/issues/728#issuecomment-909690992 и https://github.com/segmentio/kafka-go/issues/728#issuecomment-2221492034. 
Чтобы не  разработать собственный механизм взаимодействия с Schema Registry принятно решение использовать Сonfluent-kafka-go.

## Про разбиение партиций в БД 
Партиционирование таблиц в базе данных имеет смысл, если данные делятся на "горячие" и "холодные". Например, партиции можно разбивать по дате, но это не всегда отражает частоту доступа к данным.
Если в системе 100 000 000 пользователей, поиск по индексу имеет логарифмическую сложность. При 4 партициях количество шагов для поиска может снизиться с 23 до примерно 6, но прирост будет незначительным.
Еще одна идея: "холодные" данные можно переносить на более дешевые носители, чтобы они не занимали место на сервере. Однако для пользователей это нецелесообразно, так как скорость доступа должна оставаться высокой.
Также, массовую загрузку и удаление данных можно реализовать через добавление и удаление партиций, но это не применимо для сервисов регистрации.
В качестве альтернативы, можно рассмотреть "мягкое" удаление пользователей: помечать их как удалённых и перемещать в медленную партицию. Это позволит восстанавливать пользователя в течение года.
В целом, в данном случае партиционирование может быть нецелесообразным. И в результате рефакторинга партицирование по email
было удалено.

## Полезная литература

### prometheus
https://stackoverflow.com/a/65609042
https://github.com/prometheus/client_golang
https://grafana.com/oss/prometheus/exporters/go-exporter/#metrics-usage

###  otel instrumentation
https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation#new-instrumentation
https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gin-gonic/gin/otelgin/example/server.go
https://github.com/confluentinc/confluent-kafka-go/issues/712
https://github.com/etf1/opentelemetry-go-contrib

###  kafka health check
https://github.com/confluentinc/confluent-kafka-go/discussions/1041

###  using confluent-kafka-go
https://stackoverflow.com/a/55106860
https://github.com/confluentinc/confluent-kafka-go/issues/303?ysclid=lzrc7rstfd681525235#issuecomment-530566274
https://stackoverflow.com/a/69030479
https://github.com/confluentinc/confluent-kafka-go/issues/461
https://stackoverflow.com/questions/37630274/what-do-these-go-build-flags-mean-netgo-extldflags-lm-lstdc-static
https://blog.hashbangbash.com/2014/04/linking-golang-statically/ 

### kafka tracing transfer 
https://stackoverflow.com/a/78329944
https://opentelemetry.io/docs/demo/architecture/
https://github.com/open-telemetry/opentelemetry-demo/tree/e5c45b9055627795e7577c395c641f6cf240f054
https://github.com/open-telemetry/opentelemetry-demo/blob/e5c45b9055627795e7577c395c641f6cf240f054/src/checkoutservice/main.go#L527
https://www.youtube.com/watch?v=49fA7gQsDwA&t=2539s
https://www.youtube.com/watch?v=5rjTdA6BM1E
https://www.youtube.com/watch?v=UEwkn0iHDzA&list=PLNxnp_rzlqf6z1cC0IkIwp6yjsBboX945&index=1


### nginx balancer
grpc https://www.vinsguru.com/grpc-load-balancing-with-nginx/ 

```bash
curl -k --header "Content-Type: application/json" --request POST --data '{"email":"test@test.com","type":"test"}' https://localhost:443/auth/login/ 
```

```bash
curl -k --header "Content-Type: application/json" --request GET http://localhost:8090/auth/ready
```


```bash
curl -k --header "Content-Type: application/json" --request POST --data '{"email":"test@test.com","password":"test"}' http://localhost:8090/auth/registration
```

```bash
curl -k --header "Content-Type: application/json" --request POST --data '{"email":"test@test.com","password":"test"}' http://localhost:8090/auth/login
```


```bash
curl -k --header "Content-Type: application/json" --request POST --data '{"email":"test1@test.com","password":"test"}' https://localhost/auth/registration
```


```bash
curl -k --header "Content-Type: application/json" --request POST --data '{"email":"test@test.com","password":"test"}' https://localhost/auth/login
```
