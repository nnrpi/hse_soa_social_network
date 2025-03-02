# Как пользоваться:

## Sign in:
```
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "email":"test@example.com", "password":"password123"}'
```

## Log in:
```
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"password123"}' \
  -c cookies.txt
```

## Get your own profile:
```
curl -X GET http://localhost:8080/users/profile \
  -b cookies.txt
```

## Update profile:
```
curl -X PUT http://localhost:8080/users/update \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"Test", "surname":"User", "phone_number":"1234567890"}'
```

## View public profile info:
```
curl -X GET http://localhost:8080/users/testuser
```

## Log out:
```
curl -X POST http://localhost:8080/auth/logout \
  -b cookies.txt
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
INTEGRATION_TEST=true go test -v ./tests/integration/...
```
