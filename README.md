## Требования

- Docker
- Docker Compose
- Make, опционально

## Запуск
* Склонировать репозиторий
* Заполнить .env или использовать вариант по-умолчанию

```shell
cp .env.example .env
```

Запуск через Make:

```bash
make up
```

Альтернативный запуск напрямую через Docker Compose:

```bash
docker compose up -d --build
```

Сервис будет доступен на:

```text
http://localhost:8080
```

Остановка через Make:

```bash
make down
```

Остановка через Docker Compose:

```bash
docker compose down
```

Запуск тестов через make:
```bash
make test
make test-cover
```

Запуск тестов через go test
```bash
go test -v ./...
go test -v -cover ./...
```

## Примеры запросов
Создать короткую ссылку:

```bash
curl -i -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com"}'
```

Пример ответа:

```json
{
  "short_url": "http://localhost:8080/abc123"
}
```

Ожидаемый результат: `302 Found` с заголовком `Location`, указывающим на исходный URL.

## Endpoint'ы

### POST /shorten

Принимает JSON:

```json
{
  "url": "https://example.com"
}
```

Возвращает JSON:

```json
{
  "short_url": "http://localhost:8080/abc123"
}
```

### GET /{code}

Перенаправляет пользователя на оригинальный URL.

Если код не найден, возвращает `404 Not Found`.

## По проекту

`init.sql` автоматически выполняется при первом создании Docker volume для PostgreSQL.

Реализованы два endpoint'а из ТЗ:

- `POST /shorten`
- `GET /{code}`

* При повторном сокращении одной и той же ссылки сервис возвращает тот же короткий URL. То есть поведение можно считать идемпотентным
* TTL у коротких ссылок нет

- Данные БД можно посмотреть через `psql`:

```bash
docker compose exec postgres psql -U postgres -d shortener
```

## Архитектура проекта 

- `cmd/shortener/main.go` - запускает HTTP сервер, подключение к БД
- `internal/config` - чтение конфигов из env
- `internal/httpapi` - HTTP хендлеры, обработка JSON и редиректа
- `internal/shortener` - сама бизнес-логика, включет в себя: валидацию URL, генерация короткого кода и др.
- `internal/storage/postgres` - слой БД
- `init.sql` - SQL-инициализация таблицы `short_urls`.


## Конфигурация

Основные переменные окружения описаны в `.env.example`:

- `HTTP_PORT` - порт сервиса на хосте
- `HTTP_ADDR` - адрес HTTP-сервера внутри контейнера
- `BASE_URL` - базовый URL, который используется при формировании коротких ссылок
- `POSTGRES_PORT` - порт PostgreSQLы
- `POSTGRES_USER` - пользователь PostgreSQL
- `POSTGRES_PASSWORD` - пароль PostgreSQL
- `POSTGRES_DB` - имя БД
- `DATABASE_URL` - строка подключения приложения к PostgreSQL


## Ручные проверки:

- создание короткой ссылки;
- повторное создание короткой ссылки для того же URL;
- редирект по короткому коду;
- ошибка `400 Bad Request` для невалидного URL;
- ошибка `404 Not Found` для неизвестного кода;

## Некоторые особенности реализации
- Коллизии кодов обрабатываются повторной генерацией;
- Для одинакового исходного URL в БД хранится одна и таже запись, поэтому повторный запрос возвращает уже существующий короткий код;
