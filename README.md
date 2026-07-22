Ропай Анна Романовна, 225, соц сеть

# Социальная сеть

Курсовой проект по SOA — соцсеть, разбитая на несколько сервисов вместо одного монолита. Идея в том, чтобы каждый домен (пользователи, посты, статистика) жил отдельно, со своей базой, и общался с остальными либо по gRPC/REST напрямую, либо асинхронно через Kafka, когда синхронный ответ не нужен (лайки, просмотры и т.п.).

Диаграммы (C4 и ER по каждой базе) лежат в `docs/`, там же исходники в LikeC4 и PlantUML, если нужно перерисовать.

## Из чего состоит

**proxy-service** — единственная точка входа снаружи, слушает `:8080`. Сам ничего не хранит, разбирает куда роутить: `/auth` и `/users` уходят реверс-прокси на user-service, посты и промо (`/posts`, `/promos`) обрабатываются тут же через gRPC-клиент к post-service. Заодно проверяет сессию по куке `session_token`, логирует запросы, отдаёт CORS-заголовки.

**user-service** — регистрация/логин/логаут, профиль, подписки. Всё в Postgres (таблицы Users, UserProfiles, Subscriptions + сессии). При регистрации кидает событие в Kafka в топик `user_registrations`.

**post-service** (он же posts-grpc-service) — основная логика постов: создание/редактирование/удаление, лайки, комментарии, хэштеги. Тоже Postgres, тоже свои таблицы. Это gRPC-сервис, proto лежит в `src/post-service/proto`, к нему ходят и proxy, и posts-api-service. На лайк/просмотр/коммент шлёт события в Kafka.

**posts-api-service** — обёртка над posts-grpc-service, REST поверх gRPC, чтобы можно было дёрнуть curl'ом без protobuf: посмотреть пост, лайкнуть, оставить комментарий, полистать комментарии.

**statistics-service** — слушает `post_views`/`post_interactions`/`post_comments` и копит агрегированную статистику (кто лайкнул/посмотрел/прокомментировал каждый пост) в ClickHouse. Отдаёт её по `GET /stats/posts/{id}`. `user_registrations` статистикой не читается — в её ER-модели нет сущности регистраций.

## Kafka

Топики: `user_registrations`, `post_views`, `post_interactions`, `post_comments`. Брокер + zookeeper поднимаются вместе с остальным в docker-compose (образы wurstmeister), топики создаются автоматически при старте контейнера.

## Стек

Go, gorilla/mux для REST, gRPC + protobuf между сервисами, sarama как Kafka-клиент (продьюсер и консьюмер-группа), Postgres для user/post-сервисов, ClickHouse для статистики. Всё поднимается докер-компоузом, писалось и проверялось через podman, так что должно завестись и там, и там.

## Запуск

```bash
cd src
make setup        # setup_posts_services.sh, накатывает окружение
make build-posts   # генерирует proto и собирает posts-сервисы
make run           # docker-compose up
```

Дальше всё живёт на:
- proxy — localhost:8080 (сюда стучимся снаружи)
- user-service — localhost:8000
- post-service (gRPC) — localhost:9000
- posts-api-service — localhost:8090
- posts-grpc-service — localhost:9090
- statistics-service — localhost:8095
- postgres — localhost:5432
- clickhouse — localhost:8123 (HTTP), localhost:9001 (нативный протокол)
- kafka — localhost:9092

`make clean` — гасит всё и сносит volume с базой, если нужно начать с чистого листа.

## Тесты

```bash
cd src
go test ./tests/unit/...
go test ./tests/integration/...
```

Для интеграционных нужна поднятая база — гоняются отдельно от unit.

## Что уже готово, а что нет

user-service, proxy, post-service (вместе с posts-api/posts-grpc) и statistics-service реализованы и работают, покрыты тестами. `go.mod` подтягивает `clickhouse-go/v2` — перед первой сборкой прогони `go mod tidy` в `src/`, чтобы досчитались транзитивные зависимости и хэши в `go.sum`.
