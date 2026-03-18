# tramplin

API платформы Tramplin на Go + Fiber.

## Запуск через Docker Compose

Это основной способ запуска. Поднимаются:

- API
- PostgreSQL
- MinIO

```bash
cp .env.example .env
docker compose up --build
```

После старта будут доступны:

- API: `http://localhost:8081`
- Swagger UI: `http://localhost:8081/swagger/index.html`
- Swagger redirect: `http://localhost:8081/docs`
- PostgreSQL: `localhost:5433`
- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`

При запуске приложения миграции применяются автоматически.

## Миграции и PostgreSQL

После перевода репозитория на обычные PostgreSQL-таблицы приложение больше не использует `repository_state`. Если нужно вручную прогнать миграции и удалить старую таблицу состояния, используйте команды ниже.

Для Docker Compose:

```bash
docker compose down
docker compose up -d db
docker compose run --rm app goose -dir ./migrations postgres "postgres://tramplin:tramplin@db:5432/tramplin?sslmode=disable" up
docker compose up -d --build app
```

Для запуска миграций с хоста:

```bash
goose -dir ./migrations postgres "postgres://tramplin:tramplin@localhost:5433/tramplin?sslmode=disable" up
```

Проверить, что `repository_state` удалён:

```bash
psql "postgres://tramplin:tramplin@localhost:5433/tramplin?sslmode=disable" -c "select to_regclass('public.repository_state');"
```

Ожидаемый результат: `NULL`.

Проверить, что пользователи и роли лежат в обычных таблицах:

```bash
psql "postgres://tramplin:tramplin@localhost:5433/tramplin?sslmode=disable" -c "
select u.id, u.email, u.display_name, array_agg(r.code order by r.id) as roles
from users u
left join user_roles ur on ur.user_id = u.id
left join roles r on r.id = ur.role_id
group by u.id, u.email, u.display_name
order by u.created_at desc;
"
```

## Локальный запуск без Docker

Нужны:

- Go `1.25`
- PostgreSQL
- MinIO или другой S3-совместимый storage

1. Создайте `.env` на основе примера:

```bash
cp .env.example .env
```

2. Поднимите PostgreSQL и MinIO отдельно и проверьте значения переменных:

- `DATABASE_URL`
- `MIGRATIONS_DIR`
- `S3_ENDPOINT`
- `S3_ACCESS_KEY`
- `S3_SECRET_KEY`
- `S3_BUCKET`
- `S3_PUBLIC_URL`
- `JWT_SECRET`
- `JWT_TTL`

3. Запустите приложение:

```bash
go mod tidy
go run ./cmd/api
```

По умолчанию локальный API стартует на `http://localhost:8080`.

## Основные переменные окружения

Пример лежит в [.env.example](/Users/mac/Documents/codding/go/tramplin/.env.example).

Ключевые переменные:

- `APP_PORT` - порт API внутри процесса
- `DATABASE_URL` - строка подключения к PostgreSQL
- `MIGRATIONS_DIR` - путь к SQL-миграциям
- `JWT_SECRET` - секрет для подписи bearer token
- `JWT_TTL` - время жизни access token, например `24h`
- `S3_ENDPOINT` - адрес MinIO/S3
- `S3_BUCKET` - bucket для аватаров
- `S3_PUBLIC_URL` - публичный base URL для ссылок на аватары

## Swagger

Swagger генерируется из комментариев над хендлерами:

```bash
swag init -g cmd/api/main.go -o docs
```

Для защищённых ручек используйте кнопку `Authorize` и передавайте:

```text
Bearer <access_token>
```

## Аутентификация

1. Зарегистрируйтесь через `POST /api/auth/register` или войдите через `POST /api/auth/login`
2. Получите `access_token`
3. Передавайте его в заголовке:

```text
Authorization: Bearer <access_token>
```
