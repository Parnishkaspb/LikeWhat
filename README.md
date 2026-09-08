# LikeWhat gRPC service

gRPC-сервис для работы с табачными продуктами (Tobacco) и их производителями (Manufacture).

## Структура

```text
api/like_what/                     Исходный protobuf-контракт
pkg/like_what/                     Сгенерированный Go/gRPC-код
internal/likewhat/tobacco/         Домен табачных продуктов
├── tobacco.go                     Доменная модель (Tobacco, ListFilter)
├── repository/                    Интерфейс и PostgreSQL реализация (pgx)
├── service/                       Валидация и сценарии работы с табаком
└── transport/grpc/                Адаптер между gRPC и service
internal/likewhat/manufacture/     Домен производителей
├── manufacture.go                 Доменная модель (Manufacture, ListFilter)
├── repository/                    Интерфейс и PostgreSQL реализация (pgx)
├── service/                       Валидация и сценарии работы с производителями
└── transport/grpc/                Адаптер между gRPC и service
internal/platform/postgres/        Общий squirrel-builder (PostgreSQL-плейсхолдеры)
cmd/server/                        Точка входа сервера
cmd/client/                        Демонстрационный клиент
migrations/                        Миграции Goose
```

Каждый домен владеет своей моделью и не зависит от protobuf ни в сервисе, ни в репозиториях:
gRPC-контракт и прикладные правила отделены от конкретного хранилища (PostgreSQL).

## Запуск

Сервер всегда работает с PostgreSQL и требует `DATABASE_URL`:

```sh
docker compose up -d        # поднять локальную БД (или: см. ниже примечание про порт)
make migrate-up             # применить миграции
DATABASE_URL='postgres://likewhat:likewhat_dev_password@localhost:5432/likewhat?sslmode=disable' \
  go run ./cmd/server
```

Адрес gRPC-сервера меняется переменной `GRPC_ADDR` (по умолчанию `:50051`), адрес БД — `DATABASE_URL`.
In-memory хранилища нет: источник истины — база данных.

### Демонстрационный клиент

```sh
go run ./cmd/client --taste vanilla --manufacture-id <uuid>
```

## Генерация и проверка

```sh
make generate
make test
```

Тесты — интеграционные и работают на **реальной базе данных**: каждый тестовый пакет поднимает свой
PostgreSQL-контейнер через Testcontainers и применяет миграции (`internal/platform/postgres/testpostgres`).
Поэтому `go test ./...` требует доступный Docker:

- Docker Desktop — по умолчанию;
- Colima — `make test` сам подставит `DOCKER_HOST`, либо вручную:
  `export DOCKER_HOST=unix://$HOME/.colima/default/docker.sock`.

Для генерации требуются `protoc`, `protoc-gen-go` и `protoc-gen-go-grpc`. Если генераторы ещё не установлены:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

Сгенерированные файлы помещаются в `pkg/like_what` и должны обновляться при каждом изменении
`api/like_what/like_what.proto`.

## PostgreSQL для разработки

Локальная база описана в `docker-compose.yml` и запускается командой:

```sh
docker compose up -d
```

Она доступна по `localhost:5432` с БД `likewhat`, пользователем `likewhat` и паролем
`likewhat_dev_password`. Данные сохраняются в Docker volume `likewhat_postgres-data`.
Остановить контейнер можно командой `docker compose down`.

> Примечание: если порт `5432` на хосте уже занят другим PostgreSQL, измените проброс портов в
> `docker-compose.yml` (например, `"5433:5432"`) и укажите соответствующий `DATABASE_URL`.

Для применения миграций используется Goose:

```sh
make migrate-up
```

Откатить последнюю миграцию можно командой `make migrate-down`.

Миграции создают две таблицы:

- `manufactures` — производители (id, name, created_at, updated_at, deleted_at);
- `tobaccos` — табачные продукты (id, taste, photo, manufacture_id → manufactures.id, временные метки).

Идентификаторы генерируются базой (`gen_random_uuid()`). PostgreSQL-репозитории (`repository/postgres.go`)
строят параметризованные SQL-запросы через общий builder в `internal/platform/postgres` — никакой конкатенации
пользовательских данных в SQL нет.

## Архитектура

Транспорт не содержит бизнес-логики, а сервис не зависит от protobuf или конкретного хранилища. Слои:

- **transport/gRPC** — приём запросов, декодирование, нил-гард запроса и маппинг доменных ошибок в gRPC-статусы. Бизнес-валидации здесь нет;
- **service** — единственное место бизнес-валидации и нормализации данных (ozzo-validation на сервисных input-структурах), сценарии использования;
- **repository** — интерфейс хранилища и PostgreSQL-реализация `postgres.go` (on pgx).

Валидация выполняется в сервисе с помощью whitespace-aware правила `RequiredString`
(`internal/platform/validation`), чтобы отбраковать строки из одних пробелов, которые
проходят штатный `validation.Required`.
