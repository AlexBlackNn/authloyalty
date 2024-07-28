# !!!Для работы сервисов нужно запустить как сервис авторизации, так и сервис распределенных вычислений (поэтому две ссылки на 2 репозитория)

# SSO  - Сервис авторизации

Единый сервис авторизации (СА) и управления пользователями для распределенного вычислителя

## Запуск:

!При запуске скрипта прежде будут остановлены все запущенные контейнеры на ПК!  
Если есть докер контейнеры, которые не должны быть остановлены видоизмените скрипт, удалив  **docker rm -f $(docker ps -aq)** 

```bash
bash run.demo.sh 
```

## Swagger доступен по адресу:
http://127.0.0.1:44044/sw/

## Что нужно сделать 
1. Перейти на страничку со Swagger  http://127.0.0.1:44044/sw/
2. Зарегестрировать пользователя 
![reg.png](docs%2Freg.png)
3. Залогиниться и получить access и refresh токены
![login.png](docs%2Flogin.png)
4. Копируем access токен и идем на страничку Swagger распределенного вычислителя http://localhost:8080/swagger/index.html#/
5. Нажать кнопку "Authorize"
![authorize.png](docs%2Fauthorize.png)
6. Ввести access токен  (refresh токен НЕ вводить, он для обновления токенов в SSO)
![access.png](docs%2Faccess.png)
7. Вводим, интересующий запрос на вычисления и копируем id (в конкретном примере: 807fbfc0-2dec-4710-bcfa-b0c3111c8be3, может быть другим)
![calculate.png](docs%2Fcalculate.png)
8. Вводим id и получаем результат 
![result.png](docs%2Fresult.png)
9. Можно посмотреть или все запросы или отдельно запросы пользователя (выбор пользователя осуществляется на основе jwt токена)
Структура токена: 
![token_structure.png](docs%2Ftoken_structure.png)

Связь сервиса вычислителя с сервисом авторизации выполнена через grpc, взаимодействие с сервисом авторизации через  grpc. Взаимодействие с распределенным вычислителем через restapi

После запуска сервиса, можно также запустить интеграционные тесты
![tests.png](docs%2Ftests.png)

## Задачи 
1. [x] Использовать кластер PostgreSQL на базе Patroni для хранения информации о пользователях.
2. [x] Использовать кластер Redis Sentinel для сохранения отозванных токенов.
4. [x] Tесты.
5. [x] Сделать автоматический запуск кода для локальной проверки, используя Docker-compose и bash скрипты
6. [x] Связь с сервером через gRPC





#### Мысли об открытых портах в docker-compose
Не вижу смысла закрывать порты на базе, редисе и т.д. в docker-compose.
В проде врятли кто будет использовать docker-compose на 1 машине. Скорее всего 
это будет k8s или еще какой-то оркестратор, а stateful приложения, вероятно, будут
вынесены из кубера (холивар). 

Можно для  проверки подключения по grpc к сервису использовать  Postman
![postman.png](docs%2Fpostman.png)


cd authloyalty/protos/proto/registration
protoc --go_out=. registration.proto

easyjson -all /home/alex/Dev/GolandYandex/authloyalty/internal/handlers/v1/sso_handlers_response.go 


```bash
curl --header "Content-Type: application/json" --request POST --data '{"email":"test@test","password":"test"}' http://localhost:8000/auth/login
```