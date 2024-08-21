## SSO - Сервис авторизации

Важное замечание:
* Телеметрия Kafka продюсера основана на 
[opentelemetry-go-contrib](https://github.com/etf1/opentelemetry-go-contrib).
Модификации кода будут выложены в отдельный репозиторий (сейчас authloyalty/pkg/tracing/otelconfluent). Необходимость модификации вызвана
невозможностью сохранить трейс без разбиейний. Быстрый фикс. Код для обсуждения. Не для ревью 
* Тестовый консьюмер `kafka_consumer` будет удален в дальнейших версиях. Не для ревью.
* SSO - входит в общий проект по начислению балов лояльности пользователям при покупке и регистрации.

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

##### Запуск на демо стенде:
* Разворачивается инфраструктура и сервис в `docker compose`.
```bash
cd commands && task demo
```

##### Запуск интеграционных тестов:
* Разворачивается инфраструктура и сервис в `docker compose`.
* Запускаются тесты.
```bash
cd commands && task integration-tests 
```

##### Запуск юнит тестов:
* Инфраструктура НЕ разворачивается (сделаны моки).
* Запускаются тесты.
```bash
cd commands && task unit-tests 
```

### Доступ к UI:

* [Swagger grpc](http://localhost:44044/sw/)
* [Swagger http](http://localhost:8000/swagger/index.html)
* [Grafana](http://localhost:3000/grafana/) - для просмотра метрик. В папке `monitoring` лежит dashboard `6671_rev2.json`.
* [Prometheus](http://localhost:9090/targets/)
* [Kafka UI](http://localhost:8080)




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

