# Statistics Service

Слушает Kafka-топики `post_views`, `post_interactions`, `post_comments` и копит агрегированную статистику по постам в ClickHouse. Отдаёт её наружу по REST.

## Хранилище

Три таблицы (`likes`, `comments`, `views`), одна и та же схема:

```sql
CREATE TABLE IF NOT EXISTS likes (
    source_id          Int64,
    user_ids           Array(Int64),
    created_timestamp  DateTime,
    last_modified      DateTime
) ENGINE = ReplacingMergeTree(last_modified)
ORDER BY source_id;
```

На каждое событие сервис читает текущий ряд по `source_id` (через `FINAL`), добавляет `user_id` в массив и пишет новую версию строки — ClickHouse сам схлопнёт версии по `last_modified` в фоне.

## Kafka

Consumer group `statistics-service`, топики: `post_views`, `post_interactions`, `post_comments`. `user_registrations` не потребляется — в статистике нет сущности регистраций.

## REST API

- `GET /stats/posts/{id}` — счётчики по посту:
```json
{
  "post_id": 123,
  "views_count": 42,
  "unique_views_count": 30,
  "likes_count": 10,
  "unique_likes_count": 10,
  "comments_count": 5,
  "unique_comments_count": 4
}
```
Пост без единого события — валидный ответ с нулями, не 404.

## Переменные окружения

- `PORT` — порт REST API (по умолчанию 8095)
- `CLICKHOUSE_HOST`, `CLICKHOUSE_PORT`, `CLICKHOUSE_DB`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`
- `KAFKA_BROKER` (по умолчанию `kafka:9092`)
- `KAFKA_CONSUMER_GROUP` (по умолчанию `statistics-service`)
