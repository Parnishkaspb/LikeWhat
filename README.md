# LikeWhat gRPC service

gRPC-сервис для работы с табачными продуктами. Данные пока хранятся в памяти и
сбрасываются при перезапуске сервера.

## Структура

```text
api/like_what/                 Исходный protobuf-контракт
pkg/like_what/                 Сгенерированный Go/gRPC-код
internal/likewhat/tobacco/     Доменная модель
├── repository/                Интерфейс хранилища и in-memory реализация
├── service/                   Валидация и сценарии работы с табаком
└── transport/grpc/            Адаптер между gRPC и service
cmd/server/                    Точка входа сервера
cmd/client/                    Демонстрационный клиент
```

Транспорт не содержит бизнес-логики, а сервис не зависит от protobuf или
конкретного хранилища. Поэтому in-memory repository можно заменить, например,
на PostgreSQL без изменения gRPC-контракта и прикладных правил.

## Запуск

В первом терминале:

```sh
go run ./cmd/server
```

Во втором терминале:

```sh
go run ./cmd/client --taste vanilla --manufacture-id manufacturer-1
```

Адрес сервера меняется переменной `GRPC_ADDR` (по умолчанию `:50051`).

## Генерация и проверка

```sh
make generate
make test
```

Для генерации требуются `protoc`, `protoc-gen-go` и `protoc-gen-go-grpc`.
Сгенерированные файлы помещаются в `pkg/like_what` и должны обновляться при
каждом изменении `api/like_what/like_what.proto`.

## PostgreSQL для разработки

Локальная база описана в `docker-compose.yml` и запускается командой:

```sh
docker compose up -d
```

Она доступна по `localhost:5432` с базой `likewhat`, пользователем `likewhat`
и паролем `likewhat_dev_password`. Данные сохраняются в Docker volume
`likewhat_postgres-data`. Остановить контейнер можно командой
`docker compose down`.

Для применения миграций используется Goose:

```sh
make migrate-up
```

Откатить последнюю миграцию можно командой `make migrate-down`.

SQL для PostgreSQL-репозиториев строится через `github.com/Masterminds/squirrel`
с PostgreSQL-плейсхолдерами (`$1`, `$2`, …). Общий builder находится в
`internal/platform/postgres`.
