# effective-mobile-subscriptions

REST-сервис для агрегации пользовательских подписок. Приложение написано на Go и предоставляет CRUD API, аналитику суммарной стоимости подписок и Swagger-документацию. Данные сохраняются в PostgreSQL, доступна миграция для создания таблицы.

## Стек
- Go 1.22+
- PostgreSQL 14+
- Gorilla Mux
- Swagger (swaggo)
- Docker + Docker Compose

## Структура проекта
```
cmd/                    # entrypoint: инициализация, HTTP-сервер, graceful shutdown
docs/                   # swagger.json/.yaml и go-файл для генерации
internal/config/        # viper-конфиг и yaml с параметрами сервера/БД
internal/handler/       # HTTP-обработчики и swagger-комментарии
internal/service/       # бизнес-логика, валидация DTO, ошибки
internal/repository/    # работа с БД (queries, analytics)
internal/model/         # доменные структуры и DTO
migrations/             # SQL-миграции (flyway-совместимые)
```

## Настройка окружения
1. Установите Go, Docker, Docker Compose.
2. Скопируйте `.env` при необходимости (не обязателен) и отредактируйте `internal/config/config.yaml`.
3. Примените миграции (пример для Flyway или вручную через psql).

## Локальный запуск
### Через Docker Compose
```bash
docker compose up --build
```
Будут подняты сервис Go и PostgreSQL. HTTP API доступен на `http://localhost:8080`.

### Без Docker
1. Поднимите PostgreSQL и создайте базу `subscription_service`.
2. Примените `migrations/V1__create_subscriptions_table.up.sql`.
3. Настройте `internal/config/config.yaml`.
4. Запустите:
```bash
go run cmd/main.go
```

## Swagger
Документация доступна по адресу `http://localhost:8080/swagger/index.html`. Из кода генерируется пакетом swag (см. теги в хендлерах).

## Тесты
```bash
go test ./...
```

## Полезные команды
```bash
# форматирование
gofmt -w cmd/main.go internal/handler/subscription.go internal/service/subscription.go

# линтер (если подключён golangci-lint)
golangci-lint run ./...
```

## API (кратко)
- `POST /subscriptions` — создать подписку
- `GET /subscriptions` — список
- `GET /subscriptions/{id}` — получить по ID
- `PUT /subscriptions/{id}` — обновить (частично)
- `DELETE /subscriptions/{id}` — удалить
- `GET /subscriptions/analytics` — суммарная стоимость по фильтрам (`user_id`, `service_name`, `start_date_from`, `start_date_to`)

## Graceful shutdown
`cmd/main.go` использует `http.Server` с таймаутами и корректным завершением по сигналам `SIGINT/SIGTERM`, поэтому при остановке (`Ctrl+C` или `docker compose down`) текущие запросы завершаются в течение 10 секунд.


