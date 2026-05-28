# url-shortener

REST API сервис для сокращения ссылок. Поддерживает создание, редирект и удаление коротких ссылок с хранением в PostgreSQL.

---

## Содержание

- [Архитектура](#архитектура)
- [Структура проекта](#структура-проекта)
- [Требования](#требования)
- [Установка и запуск](#установка-и-запуск)
- [Конфигурация](#конфигурация)
- [API](#api)
- [CLI](#cli)
- [Тестирование](#тестирование)

---

## Архитектура

Проект состоит из двух исполняемых файлов:

- **server** - HTTP-сервер, обрабатывает запросы на создание, редирект и удаление ссылок.
- **cli** - консольный клиент, предоставляет интерактивный интерфейс для взаимодействия с сервером.

Оба бинарника собираются из соответствующих `main.go` и работают независимо.

### Слои приложения

```
cmd/
├── server/main.go     — точка входа сервера (роутинг, middleware, graceful shutdown)
└── cli/main.go        — точка входа CLI-клиента

internal/
├── config/            — загрузка конфигурации из YAML + .env
├── logger/            — настройка slog (text / JSON / prod-JSON в зависимости от окружения)
├── storage/           — работа с PostgreSQL (CRUD: SaveUrl, GetUrl, DeleteUrl, GetInfo)
├── handlers/
│   ├── alias/         — POST /url — создание короткой ссылки
│   ├── redirect/      — GET /{alias} — редирект на оригинальный URL
│   ├── deleter/       — DELETE /url/{alias} — удаление ссылки
│   └── api/           — общий JSONHandler для формирования ответов
└── middleware/
    ├── Auth.go        — Bearer-токен авторизация для защищённых маршрутов
    ├── logger.go      — логирование каждого входящего запроса через slog
    └── ratelimiter.go — ограничение частоты запросов по IP (500 мс между запросами)

lib/
└── random/            — генерация случайного алиаса (да, тут я использую math/rand, а не crypto)

local/
└── local.yaml         — конфигурация для локального окружения
```

---

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/<username>/url-shortener.git
cd url-shortener
```

### 2. Переменные окружения

Скопируйте `.env.example` и заполните значения:

```bash
cp .env.example .env
```

```env
DB_CONN="host=localhost port=5432 user=USER password=PASSWORD dbname=NAME sslmode=disable"
CONFIG_PATH=./local/local.yaml
MIDDLEWARE_TOKEN="YOUR_SECRET_TOKEN"
```

- `DB_CONN` — строка подключения к PostgreSQL.
- `CONFIG_PATH` — путь до YAML-конфига. Если не задан, используется `./local/local.yaml`.
- `MIDDLEWARE_TOKEN` — токен для авторизации DELETE-запросов.

### 3. Сборка

```bash
# Сервер
go build -o server.exe ./cmd/main

# CLI-клиент
go build -o cli.exe ./cmd/cli
```

### 4. Запуск

Сначала запустите сервер:

```bash
./server.exe
```

Затем в отдельном терминале запустите CLI:

```bash
./cli.exe
```

Сервер поднимается на адресе, указанном в `local.yaml` (по умолчанию `localhost:8082`). При старте автоматически создаётся таблица в базе данных, если она ещё не существует.

---


## API

### Создать короткую ссылку

```
POST /url
Content-Type: application/json
```

**Тело запроса:**

```json
{
  "url": "https://example.com/very/long/path",
  "alias": "mylink"
}
```

Поле `alias` необязательно. Если не указано, генерируется случайная строка длиной 6 символов.

**Ответ `200 OK`:**

```json
{
  "original_url": "https://example.com/very/long/path",
  "alias": "mylink",
  "RespStatus": {
    "status": "OK"
  }
}
```

---

### Редирект по алиасу

```
GET /{alias}
```

Возвращает `302 Found` с заголовком `Location`, указывающим на оригинальный URL. При каждом обращении счётчик переходов (`clicks`) увеличивается на 1.

---

### Удалить ссылку

```
DELETE /url/{alias}
Authorization: Bearer <MIDDLEWARE_TOKEN>
```

Требует заголовок `Authorization`. Возвращает `200 OK` при успехе.

---

## CLI

CLI предоставляет интерактивное меню в терминале:

```
*    Сокращатель ссылок    *
Введите 'exit' для выхода

1. Создать ссылку
2. Удалить ссылку
```

- **Создать ссылку** — запрашивает URL и (опционально) алиас, отправляет `POST /url`.
- **Удалить ссылку** — запрашивает алиас, отправляет `DELETE /url/{alias}` с токеном из `.env`.
- **exit** — завершает работу.

CLI берёт `MIDDLEWARE_TOKEN` из того же `.env`, что и сервер.

---

## Тестирование

Юнит-тесты написаны для хендлеров `redirect` и `deleter` с использованием mock-интерфейсов.

```bash
go test ./...
```


**redirect:**
- успешный редирект → `302 Found` + корректный `Location`
- алиас не найден → `404 Not Found`
- пустой алиас → `400 Bad Request`

**deleter:**
- успешное удаление → `200 OK`
- ошибка базы данных → `500 Internal Server Error`
- пустой алиас → `400 Bad Request`
- корректная передача алиаса в слой хранилища