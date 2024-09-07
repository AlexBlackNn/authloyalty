# Сервис начисления баллов лояльности с авторизацией 

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

### Локальный запуск:
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

3. Запуск сервиса начисления баллов лояльности
```bash
go run ./loyalty/cmd/main.go --config=./loyalty/config/local.yaml
```

4. Описание хэндлеров 
 4.1 sso
   В сервисе авторизации также доступны handlers для:
   1. Login - получить токены доступа
   2. Logout - отозвать токен
   3. Refresh - обновить токен (необходимо использовать refresh token)
   4. Register - зарегестировать пользователя

  4.2 loyalty 
  1. AddLoyalty - начислить или списать баллы
  2. GetLoyalty - получить баллы


5. Взаимодействие между сервисами
    
    4.1. Открыть swagger   
    a. [swagger](http://localhost:8000/swagger/index.htm/index.html) сервиса авторизации
    
    б. [swagger](http://localhost:8001/swagger/index.htm/index.html) сервиса начисления баллов лояльности 

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
   
    4.3 В swagger сервиса начисления баллов лояльности нажать Authorize и ввести access_token из пункта 4.2
    4.4 Списание баллов лояльности
 
   ![loyalty_withdraw.png](docs%2Floyalty_withdraw.png)

   uuid из тела запроса, анализируется, только если запрос пришел от аккаунта администратора (jwt token содержит поле admin)  : 
   Пример тела запроса:
   ```
   {
   "balance": 20,
   "comment": "purchase",
   "operation": "w",
   "uuid": "7b4825bd-1c03-43ed-9470-3906015b6fc0"
   }
   ``` 
   В случае пользователя uuid извлекается из jwt token. 
   Операция начисления баллов, доступна только администратору или при регистрации пользователя (приходит сообщение по шине данных kafka).
   
   ###  Демо запуск: 

Поднимются все сервисы в докере. Так как сервис авторизации находится за балансировщиками нагрузки, подымаются его 2 копии (можно больше, трафик будет случайным образом (round robin) перенаправляться в одну из копий).

```bash
cd commands && task demo-scale
```

#### sso - Примеры запросов (необходимо подстроить под себя, так как токены могут отличаться)
1. Регистрация
```bash
curl -k -X 'POST' \
  'https://localhost:443/auth/registration' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "birthday": "string",
  "email": "test00@test.com",
  "name": "test_name",
  "password": "test"
}'
```

Результат: 
```
{
"status":"Success",
"user_id":"e66a298a-c835-4cfa-8c8e-60534f4c00f9","access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjUxOTM4MTksInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiJlNjZhMjk4YS1jODM1LTRjZmEtOGM4ZS02MDUzNGY0YzAwZjkifQ.oL2ndWYnf6I4vC9xThtCLsFVyqSje3a__n1Iz_sYC6g","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjYwNTQyMTksInRva2VuX3R5cGUiOiJyZWZyZXNoIiwidWlkIjoiZTY2YTI5OGEtYzgzNS00Y2ZhLThjOGUtNjA1MzRmNGMwMGY5In0.FecpRypAwL88rNy4HPe3pyWFNWnmEq71r1Cae5rbpZ0"
}
```

2. Логин
```bash
curl -k -X 'POST' \
  'https://localhost:443/auth/login' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "email": "test@test.com",
  "password": "test"
}'
```

Результат
```
{
"status":"Success",
"user_id":"e66a298a-c835-4cfa-8c8e-60534f4c00f9",
"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjUxOTM5MTEsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiJlNjZhMjk4YS1jODM1LTRjZmEtOGM4ZS02MDUzNGY0YzAwZjkifQ.D7SQF4GlM3ykCh8gZbb6qUOAxufbha_0ZEW1GEsei2g",
"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjYwNTQzMTEsInRva2VuX3R5cGUiOiJyZWZyZXNoIiwidWlkIjoiZTY2YTI5OGEtYzgzNS00Y2ZhLThjOGUtNjA1MzRmNGMwMGY5In0.xuQlt_TWB1y8vLHJl-2fsmP5ICnDOXFmA1YIryp2X_c"
}
```

3. Отзыв токена 
```bash
   curl -k -X 'POST' \
   'https://localhost:443/auth/logout' \
   -H 'accept: application/json' \
   -H 'Content-Type: application/json' \
   -d '{
   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjUxOTM5MTEsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiJlNjZhMjk4YS1jODM1LTRjZmEtOGM4ZS02MDUzNGY0YzAwZjkifQ.D7SQF4GlM3ykCh8gZbb6qUOAxufbha_0ZEW1GEsei2g"
   }'
```

Результат: 
```
{
"status":"Success"
}
```

4. Обновление токена 
```bash
curl -k -X 'POST' \
  'https://localhost:443/auth/refresh' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjYwNTQzMTEsInRva2VuX3R5cGUiOiJyZWZyZXNoIiwidWlkIjoiZTY2YTI5OGEtYzgzNS00Y2ZhLThjOGUtNjA1MzRmNGMwMGY5In0.xuQlt_TWB1y8vLHJl-2fsmP5ICnDOXFmA1YIryp2X_c"
}'
```

Результат: 
```
{
"status":"Success",
"user_id":"e66a298a-c835-4cfa-8c8e-60534f4c00f9",
"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjUxOTQzOTgsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiJlNjZhMjk4YS1jODM1LTRjZmEtOGM4ZS02MDUzNGY0YzAwZjkifQ.JGhm4XYUHWEasWdcHZWkGyxRtMg7CbldLvtlGKd-tWA",
"refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjYwNTQ3OTgsInRva2VuX3R5cGUiOiJyZWZyZXNoIiwidWlkIjoiZTY2YTI5OGEtYzgzNS00Y2ZhLThjOGUtNjA1MzRmNGMwMGY5In0.IxZWJD3Ry6-uL992nWFcfqvqnQ-uGguXXTTHC9fvoe8"
}
```

#### loyalty - Примеры запросов (необходимо подстроить под себя, так как токены могут отличаться)

1. Изменение баллов лояльности
```bash
curl -k -X 'POST' \
'http://localhost:8001/loyalty' \
-H 'accept: application/json' \
-H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAdGVzdC5jb20iLCJleHAiOjE3MjUxOTU3MjUsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJ1aWQiOiJlNjZhMjk4YS1jODM1LTRjZmEtOGM4ZS02MDUzNGY0YzAwZjkifQ.6k9LyerCfpTcEqv4bDY2CLxPWuhx-JZ2pi2Ew3tBx84' \
-H 'Content-Type: application/json' \
-d '{
"balance": 1,
"comment": "purchase",
"operation": "w",
"uuid": "e66a298a-c835-4cfa-8c8e-60534f4c00f9"
}'
```

2. Получить кол-во баллов лояльности по uuid пользователя
```bash
curl -k -X 'GET' \
'https://localhost:443/loyalty/e66a298a-c835-4cfa-8c8e-60534f4c00f9'
```

Результат:
```
{"status":"Success","uuid":"e66a298a-c835-4cfa-8c8e-60534f4c00f9","balance":78}
```

3. Ready
```bash
curl -k -X 'GET' \
'https://localhost:443/loyalty/ready'
```


## Архитектурные решения

![architecture.jpg](docs%2Farchitecture.jpg)

### API:
HTTP-handlers используются для общения с frontend, а gRPC для общения внутри микросервисов.
* gRPC: использует Protobuf для сериализации данных, что обеспечивает более компактное
  представление сообщения по сравнению с JSON (используемым в HTTP), следовательно
  реализуется более быстрый обмен данными.
* HTTP: стандарт для веб-разработки, что делает его простым и легким для
  интеграции с frontend-фреймворками.

### API Gateway
программный слой, который выступает в качестве единого входной точки для фронтенда. Он перенаправляет запросы к соответствующим сервисам.

### Регистрация пользователей

Система реализует немедленную выдачу токенов при регистрации(только в http API, в случае grpc handler
регистрации будет удален), позволяя пользователю начать работу с сервисом сразу. Для обеспечения высокой доступности данных используется Patroni кластер с асинхронной репликацией, что может приводить к кратковременной несогласованности данных между ведущим узлом и репликами. Однако, время жизни токенов (15 минут) достаточно для завершения репликации. Чтение данных с реплик снижает нагрузку на ведущий узел, повышая производительность системы.

Несмотря на риск временной несогласованности, этот подход обеспечивает баланс между производительностью и согласованностью данных. Пользователь может столкнуться с кратковременными проблемами авторизации в случае больших нагрузок или сетевых проблем, но вероятность этого мала.

### Асинхронная обработка событий регистрации

Сервис использует Kafka для асинхронной обработки событий регистрации пользователей.
При успешной регистрации генерируется сообщение в Kafka-топик, содержащее информацию о новом пользователе. Сервис отслеживает статус доставки сообщений и обновляет его в базе данных.
Это позволяет другим сервисам, например, сервису отправки приветственных сообщений или начисления баллов лояльности, подписываться на этот топик и асинхронно обрабатывать информацию о новых пользователях.

Каждый из сервисов работает независимо, и у него есть своя база данных. Возникают распределённые транзакции, а для управления ими используется паттерн Saga.  Транзакциями управляют через оркестрацию или хореографию. В качестве реалзиации управления транзакциями выбрана хореография,
так как взаимодействие между микросервисами в проекте простое и не требует сложной координации.В этом подходе нет центрального управляющего компонента. Каждый микросервис самостоятелен — он знает, что
делать после выполнения своего шага. Микросервисы взаимодействуют через события и сами инициируют компенсирующие действия. 

Обработка ошибок играет важную роль. Каждый шаг транзакции выполняют разные микросервисы, поэтому ошибки могут возникнуть на любом из этапов.

В качестве обработки ошибок используются:
1. локальная отмена транзакций 
2. повторные попытки (например, если при регистрации пользователя возникла ошибка отправки сообщения в брокер сообщений, то такое сообщение помечается в БД как звершенное с ошибкой. В дальнейшем необходимо сделать автоматическую переотпавку таких сообщений )
3. тайм-ауты и дедлайны
4. системы мониторинга и алертов (логи: Promteil, Loki, Grafana; трасировка: Jaeger, метрики: Prometheus).


### Выбор библиотеки для взаимодействия сервисом через Kafka

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

### Про разбиение партиций в БД
Партиционирование таблиц в базе данных имеет смысл, если данные делятся на "горячие" и "холодные". Например, партиции можно разбивать по дате, но это не всегда отражает частоту доступа к данным.
Если в системе 100 000 000 пользователей, поиск по индексу имеет логарифмическую сложность. При 4 партициях количество шагов для поиска может снизиться с 23 до примерно 6, но прирост будет незначительным.
Еще одна идея: "холодные" данные можно переносить на более дешевые носители, чтобы они не занимали место на сервере. Однако для пользователей это нецелесообразно, так как скорость доступа должна оставаться высокой.
Также, массовую загрузку и удаление данных можно реализовать через добавление и удаление партиций, но это не применимо для сервисов регистрации.
В качестве альтернативы, можно рассмотреть "мягкое" удаление пользователей: помечать их как удалённых и перемещать в медленную партицию. Это позволит восстанавливать пользователя в течение года.
В целом, в данном случае партиционирование может быть нецелесообразным. И в результате рефакторинга партицирование по email было удалено.


## Что еще требуется сделать

1. В кафке убрать передачу email и в токенах тоже (так как это приватная инфа)
2. Сделать взаимодействие между сервисами, чтобы 1 мог спросить у другого инфу о пользователе (нужно будет для сервисов логики получения токенов сделать)
3. Разделить ручки в сервисе лояльности для админов и клиентов.
4. Добавить больше тестов, в том числе и интеграциооные на несколько сервисов
5. Написать CI/CD