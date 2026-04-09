# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Допущения

1. Задача может быть обычной или периодической.

2. Для задачи вводится дата выполнения `scheduled_at`.

3. Настройки повторяемости хранятся у самой задачи.

4. Поддерживаются типы повторяемости `daily`, `monthly`, `specific_dates`, `day_parity`.

5. Для `monthly` день месяца ограничен значениями от 1 до 30.

6. Генерация бесконечного числа будущих экземпляров задач не реализуется. Сервис хранит правило повторяемости. Это правило может использовать отдельный планировщик в будущем.

## Что реализовано

Добавлено поле `scheduled_at` для даты выполнения задачи.

Добавлена поддержка периодических задач через настройки `recurrence`.

Поддержаны типы повторяемости:
1. `daily`
2. `monthly`
3. `specific_dates`
4. `day_parity`

Добавлена валидация правил повторяемости на уровне бизнес логики.

Расширены создание, чтение и обновление задач с учетом новых полей.

## Проверенные сценарии

1. Создание обычной задачи с `scheduled_at`.
2. Создание задачи с `recurrence` типа `daily`.
3. Создание задачи с `recurrence` типа `monthly`.
4. Создание задачи с `recurrence` типа `specific_dates`.
5. Создание задачи с `recurrence` типа `day_parity`.
6. Получение списка задач с новыми полями.
7. Получение задачи по id с новыми полями.
8. Обновление задачи с новыми полями.
9. Ошибка валидации при `monthly`, если `day_of_month` больше 30.