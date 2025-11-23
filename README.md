# PR Reviewer Assignment Service

Краткое руководство, чтобы быстро запустить проект в Docker.

---

## 🚀 Быстрый запуск с Docker (рекомендуется)

Запустить сервис и БД:

```bash
docker-compose up -d --build
```

Проверить логи:

```bash
docker logs -f pr-service      # логи приложения
docker logs -f pr-db           # логи Postgres
```

Остановить:

```bash
docker-compose down
```

---

## 📡 API Endpoints (кратко)

| Method | Path                        | Description                           |
|--------|----------------------------|---------------------------------------|
| GET    | /health                    | Health check                           |
| POST   | /team/add                  | Создать / обновить команду             |
| GET    | /team/get                  | Получить команду                       |
| POST   | /team/deactivate           | Массовая деактивация                   |
| POST   | /users/setIsActive         | Активировать / деактивировать пользователя |
| GET    | /users/getReview           | PR, где он ревьювер                    |
| POST   | /pullRequest/create        | Создать PR + автоназначение            |
| POST   | /pullRequest/merge         | Слить PR (идемпотентно)               |
| POST   | /pullRequest/reassign      | Переназначить ревьювера               |
| GET    | /stats/reviewers           | Статистика по ревьюверам              |

---

## 🧪 Примеры CURL-запросов

### Создать команду

```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name":"backend",
    "members":[
      {"user_id":"u1","username":"Alice","is_active":true},
      {"user_id":"u2","username":"Bob","is_active":true},
      {"user_id":"u3","username":"Charlie","is_active":true}
    ]
  }'
```

### Создать PR

```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id":"pr-1","pull_request_name":"Fix bug","author_id":"u1"}'
```

### Получить PR текущего ревьювера

```bash
curl "http://localhost:8080/users/getReview?user_id=u2"
```

### Массовая деактивация

```bash
curl -X POST http://localhost:8080/team/deactivate \
  -H "Content-Type: application/json" \
  -d '{"team_name":"backend"}'
```

### Просмотр статистики

```bash
curl http://localhost:8080/stats/reviewers
```
