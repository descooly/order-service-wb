# Order Service — демонстрационный сервис заказов

Демонстрационный сервис для приёма заказов через **NATS Streaming**, сохранения в **PostgreSQL** и отдачи данных по **HTTP**.  
Включает **in-memory кэш**, **восстановление из БД при старте**, **graceful shutdown** и простой **HTML-интерфейс** для просмотра заказов.

## Используемые технологии

- **Golang** — основной бэкенд (модули, unit-тесты)
- **PostgreSQL** — хранилище данных
- **NATS Streaming** — брокер сообщений
- **Docker & Docker Compose** — оркестрация
- **HTML + CSS** — минималистичный веб-интерфейс
- **golang-migrate** — управление миграциями базы данных

## Основные функции

- Подписка на канал `orders` с **durable subscription** и ручным подтверждением (`msg.Ack()`).
- Валидация входящих сообщений: проверка на валидный JSON и обязательное поле `order_uid`.
- Запись данных в PostgreSQL с атомарной транзакцией по связанным таблицам:
  - `order_info`
  - `delivery`
  - `payment`
  - `items`
- In-memory кэш на основе `map[string]OrderStruct` с потокобезопасным доступом (`sync.RWMutex`).
- Восстановление кэша из БД при запуске сервиса.
- HTTP API:
  - `GET /order?order_uid=...` — возвращает заказ из кэша в формате JSON.
- Простой веб-интерфейс (`/`) с формой ввода `order_uid`.
- Graceful shutdown: корректное завершение при `Ctrl+C` (ожидание завершения обработки, закрытие соединений).
- Управление схемой БД через **SQL-миграции** (инструмент `golang-migrate`).
- Статические файлы встроены в бинарник с помощью `//go:embed`.

## Структура проекта

```
project/
├── internal/
│   ├── cache/              # In-memory кэш
│   │   └── cache.go        # OrderCache + Set/Get
│   ├── db/                 # Работа с БД
│   │   └── dbConnect.go    # Подключение, InsertOrder, LoadOrders
│   ├── httpserver/         # HTTP-сервер и обработчики
│   │   ├── server.go       # Запуск сервера, маршруты
│   │   └── static/         # Веб-интерфейс (встроен через embed)
│   ├── model/              # Структуры данных
│   │   └── struct.go       # OrderStruct, Delivery, Payment, Item
│   ├── nats/               # Работа с NATS Streaming
│   │   └── subscriber.go   # Подписка на канал
│   └── service/            # Бизнес-логика
│       └── service.go      # Обработка сообщений и HTTP-запросов
├── main/                   # Точка входа
│   └── main.go             # Инициализация зависимостей и запуск
├── migrations/             # SQL-миграции
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── publisher/              # Утилита для публикации тестовых заказов
│   └── publisher.go
├── .dockerignore           # Исключение файлов из Docker-образа
├── .env                    # Переменные окружения для Docker Compose
├── docker-compose.yml      # Оркестрация контейнеров
├── Dockerfile              # Сборка Go-образа
├── go.mod
└── go.sum
```

## Как запустить проект

### Требования

- Docker и Docker Compose

### Запуск через Docker Compose

```bash
docker-compose up --build
```

Это:

- Создаст и запустит PostgreSQL
- Применит SQL-миграции из папки `migrations/`
- Запустит NATS Streaming
- Соберёт и запустит Go-сервис
- Раздаст статику через встроенный HTTP-сервер

Откройте в браузере: http://localhost:8080

## Как протестировать

### Отправьте тестовый заказ:

```bash
cd publisher
go run main.go
```

Сервис получит сообщение, сохранит в БД и кэш.

### Проверьте через веб-интерфейс:

- Перейдите на http://localhost:8080
- Введите `b563feb7b2b84b6test`
- Вы увидите JSON с деталями заказа

## Стресс-тесты

Тестирование выполнено с помощью `vegeta`.

**vegeta (rate=100, 30s):**
```
Requests      [total, rate, throughput]         3000, 100.04, 100.04
Duration      [total, attack, wait]             29.989s, 29.989s, 0s
Latencies     [min, mean, 50, 90, 95, 99, max]  0s, 194.911µs, 0s, 560.974µs, 600.693µs, 1.19ms, 20.333ms
Bytes In      [total, mean]                     1998000, 666.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:3000
```

**vegeta (rate=500, 30s):**
```
Requests      [total, rate, throughput]         15000, 500.04, 500.04
Duration      [total, attack, wait]             29.997s, 29.997s, 0s
Latencies     [min, mean, 50, 90, 95, 99, max]  0s, 79.591µs, 0s, 118.788µs, 550.994µs, 1.122ms, 22.501ms
Bytes In      [total, mean]                     9990000, 666.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:15000
```

**vegeta (rate=1000, 10s):**
```
Requests      [total, rate, throughput]         10000, 1000.14, 1000.14
Duration      [total, attack, wait]             9.999s, 9.999s, 0s
Latencies     [min, mean, 50, 90, 95, 99, max]  0s, 111.538µs, 0s, 36.432µs, 141.1µs, 2.427ms, 28.144ms
Bytes In      [total, mean]                     6660000, 666.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:10000
```

**vegeta (rate=10000, 10s):**
```
Requests      [total, rate, throughput]         99999, 10000.11, 9803.71
Duration      [total, attack, wait]             10s, 10s, 0s
Latencies     [min, mean, 50, 90, 95, 99, max]  0s, 9.426ms, 0s, 845.471µs, 73.192ms, 237.381ms, 494.721ms
Bytes In      [total, mean]                     65291310, 652.92
Success       [ratio]                           98.04%
Status Codes  [code:count]                      0:1964  200:98035
Error Set:
Get "http://localhost:8080/order?order_uid=TEST123": dial tcp ...: connectex: No connection could be made because the target machine actively refused it.
```

> Примечание: при высокой нагрузке (10k RPS) возможны временные отказы соединения из-за ограничений ОС или Docker. Сервис остаётся стабильным при нагрузке до 1000 RPS.

## Тестирование

### Запуск unit-тестов:

```bash
go test ./...
```

Покрыты:
- Кэш (`internal/cache`)
- HTTP-обработчик (`main/main_test.go`)

### Graceful Shutdown

При получении сигнала `SIGINT` или `SIGTERM`:
- Останавливается подписка на NATS
- Закрывается соединение с PostgreSQL
- Корректно завершается HTTP-сервер

## Инициализация базы данных

Схема базы данных управляется через **миграции**.  
При первом запуске `docker-compose` автоматически применяет все миграции из папки `migrations/`.  
Файл `init.sql` больше не используется.
---

## Короткая демонстрация работы 

![document_5469637832094023159 (1)](https://github.com/user-attachments/assets/00c03838-857f-4cc6-bfaa-231d4a440885)
