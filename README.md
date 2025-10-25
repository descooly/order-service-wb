# Order Service — демонстрационный сервис заказов

Демонстрационный сервис для приёма заказов через **NATS Streaming**, сохранения в **PostgreSQL** и отдачи данных по **HTTP**.  
Включает **in-memory кэш**, **восстановление из БД при старте**, **graceful shutdown** и простой **HTML-интерфейс** для просмотра заказов.

---

## Используемые технологии

- **Golang** — основной бэкенд (модули, unit-тесты)
- **PostgreSQL** — хранилище данных
- **NATS Streaming** — брокер сообщений
- **Docker & Docker Compose** — оркестрация
- **HTML + CSS** — минималистичный веб-интерфейс

---

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

---

## Структура проекта

```
project/
├── internal/
│   ├── cache/              # In-memory кэш
│   │   └── cache.go        # OrderCache + Set/Get
│   └── db/                 # Работа с БД
│       ├── dbConnect.go    # Подключение к PostgreSQL, InsertOrder, LoadOrders
│       └── struct.go       # Структуры OrderStruct, Delivery, Payment, Item
├── publisher/              # Утилита для публикации тестовых заказов
│   └── publisher.go
├── static/                 # Статические файлы
│   ├── index.html          # Главная страница
│   └── style.css           # Стили
├── main.go                 # Точка входа (NATS, HTTP, graceful shutdown)
├── init.sql                # SQL-скрипт создания таблиц
├── docker-compose.yml      # Оркестрация контейнеров
├── Dockerfile              # Сборка Go-образа
├── go.mod
└── go.sum
```

---

## Как запустить проект

### Требования

- Docker и Docker Compose

---

### Запуск через Docker Compose

```bash
docker-compose up --build
```

Это:

1. Создаст и запустит PostgreSQL с таблицами из `init.sql`
2. Запустит NATS Streaming
3. Соберёт и запустит Go-сервис
4. Раздаст статику через встроенный HTTP-сервер

Откройте в браузере: [http://localhost:8080](http://localhost:8080)

---

### Как протестировать

1. **Отправьте тестовый заказ**:
   ```bash
   cd publisher
   go run main.go
   ```
   Сервис получит сообщение, сохранит в БД и кэш.

2. **Проверьте через веб-интерфейс**:
   - Перейдите на [http://localhost:8080](http://localhost:8080)
   - Введите `b563feb7b2b84b6test`
   - Вы увидите JSON с деталями заказа

---

## Стресс-тесты

> Результаты тестов 

### vegeta (rate=100, 30s):
```
Requests      [total, rate, throughput]         3000, 100.03, 100.03
Duration      [total, attack, wait]             29.99s, 29.99s, 535.6µs
Latencies     [mean, 50, 95, 99, max]          417.259µs, 541.416µs, 868.268µs, 1.033507ms, 27.1522ms
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:3000
```

### vegeta (rate=500, 30s):
```
Requests      [total, rate, throughput]         15000, 500.03, 500.03
Duration      [total, attack, wait]             29.997s, 29.997s, 0s
Latencies     [mean, 50, 95, 99, max]          57.433µs, 0s, 336.615µs, 575.18µs, 21.4045ms
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:15000
```

### vegeta (rate=1000, 10s):
```
Requests      [total, rate, throughput]         10000, 1000.11, 1000.11
Duration      [total, attack, wait]             9.998s, 9.998s, 0s
Latencies     [mean, 50, 95, 99, max]          50.073µs, 0s, 80.896µs, 480.599µs, 23.5528ms
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:10000
```

### vegeta (rate=10000, 10s):
```
Requests      [total, rate, throughput]         99995, 10000.24, 10000.24
Duration      [total, attack, wait]             9.999s, 9.999s, 0s
Latencies     [mean, 50, 95, 99, max]          245.938µs, 0s, 582.768µs, 2.239727ms, 90.3543ms
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:99995
```

---

## Тестирование

Запуск unit-тестов:
```bash
go test ./...
```

Покрыты:
- Кэш (`internal/cache`)
- HTTP-обработчик (`main_test.go`)

---

## Graceful Shutdown

При получении сигнала `SIGINT` или `SIGTERM`:
- Останавливается подписка на NATS
- Закрывается соединение с PostgreSQL
- Корректно завершается HTTP-сервер

---

## Инициализация базы данных

Файл `init.sql` содержит `CREATE TABLE` для всех необходимых таблиц.  
При первом запуске `docker-compose` автоматически выполнит этот скрипт.

---

## Короткая демонстрация работы 

![document_5469637832094023159 (1)](https://github.com/user-attachments/assets/00c03838-857f-4cc6-bfaa-231d4a440885)
