# Запуск GophKeeper с Docker Compose

## Требования

- Docker
- Docker Compose

## Быстрый старт

1. Убедитесь, что файл `.env` существует и содержит правильные настройки:
   ```bash
   cp .env_example .env
   ```

2. Запустите все сервисы:
   ```bash
   docker-compose up -d
   ```

3. Проверьте статус сервисов:
   ```bash
   docker-compose ps
   ```

4. Посмотрите логи сервера:
   ```bash
   docker-compose logs -f server
   ```

## Сервисы

Docker Compose запускает три сервиса:

- **postgres**: PostgreSQL 16 (порт 5432)
- **redis**: Redis 7 (порт 6379)
- **server**: GophKeeper gRPC сервер (порт 50051)

## Управление

### Остановка сервисов
```bash
docker-compose down
```

### Остановка с удалением volumes
```bash
docker-compose down -v
```

### Пересборка сервера
```bash
docker-compose up -d --build server
```

### Перезапуск сервера
```bash
docker-compose restart server
```

## Применение миграций

Миграции применяются автоматически при запуске сервера. Если нужно применить миграции вручную:

```bash
docker-compose exec server migrate -path /app/migrations -database "$DATABASE_DSN" up
```

## Подключение к базе данных

Для подключения к PostgreSQL из контейнера:

```bash
docker-compose exec postgres psql -U postgres -d gophkeeper
```

Для подключения с хоста:
```bash
psql -h localhost -p 5432 -U postgres -d gophkeeper
```

## Подключение к Redis

Для подключения к Redis из контейнера:

```bash
docker-compose exec redis redis-cli
```

Для подключения с хоста:
```bash
redis-cli -h localhost -p 6379
```

## Просмотр логов

### Все сервисы
```bash
docker-compose logs -f
```

### Конкретный сервис
```bash
docker-compose logs -f server
docker-compose logs -f postgres
docker-compose logs -f redis
```

## Troubleshooting

### Сервер не запускается

1. Проверьте логи:
   ```bash
   docker-compose logs server
   ```

2. Убедитесь, что PostgreSQL и Redis запущены:
   ```bash
   docker-compose ps
   ```

3. Проверьте подключение к базе данных:
   ```bash
   docker-compose exec postgres pg_isready -U postgres
   ```

### Проблемы с миграциями

Если миграции не применяются, попробуйте:
```bash
docker-compose exec server migrate -path /app/migrations -database "$DATABASE_DSN" force 0
docker-compose exec server migrate -path /app/migrations -database "$DATABASE_DSN" up
```

### Полная очистка

Для удаления всех данных и перезапуска с нуля:
```bash
docker-compose down -v
docker-compose up -d
```

## Разработка

Для разработки с горячей перезагрузкой можно использовать локальный запуск сервера, а базы данных запускать через Docker:

```bash
# Запустить только базы данных
docker-compose up -d postgres redis

# Запустить сервер локально
make run
```

В этом случае в `.env` файле используйте `localhost` вместо имен сервисов:
```
DATABASE_DSN=postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable
REDIS_ADDR=localhost:6379