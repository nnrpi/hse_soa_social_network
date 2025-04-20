# Как пользоваться:

## Sign in:
```
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"password123", "email":"user@example.com"}'
```

## Log in:
```
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"username":"testuser", "password":"password123"}'
```

## Create post:
```
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"title":"Мой первый пост", "description":"Это описание моего первого поста", "is_private":false, "tags":["тест", "первый"]}'
```

## Get post by id:
```
curl -X GET http://localhost:8080/posts/{id} -b cookies.txt
```

## Update post:
```
curl -X PUT http://localhost:8080/posts/{id} \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"title":"Обновленный пост", "description":"Новое описание поста", "is_private":true, "tags":["обновлено", "тест"]}'
```

## Get post list:
```
curl -X GET "http://localhost:8080/posts?page=1&page_size=10" -b cookies.txt
```

## Delete post:
```
curl -X DELETE http://localhost:8080/posts/{id} -b cookies.txt
```

## Log out:
```
curl -X POST http://localhost:8080/auth/logout -b cookies.txt
```

Везде вместо ``-b cookies.txt`` можно писать ``-H "Cookie: session_token=<session-token>"``


# Как запускать тесты:

## Unit tests:
```
go test -v ./tests/unit/...
```

## Integration tests:

```
podman run --name postgres-test -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=socialnetwork_test -p 5432:5432 -d postgres:14
INTEGRATION_TEST=true go test -v ./tests/integration/post_tests/... -count=1
```
