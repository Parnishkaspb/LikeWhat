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
