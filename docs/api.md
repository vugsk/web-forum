# REST API

## Обзор

REST API v1 предназначен для мобильных приложений и сторонних интеграций.

**Базовый URL:** `/api/v1`

## Формат ответов

### Успешный ответ

```json
{
  "success": true,
  "data": { ... }
}
```

### Ошибка

```json
{
  "success": false,
  "error": "Описание ошибки"
}
```

## CORS

API поддерживает CORS для всех origins (для разработки):

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## Доски

### Получить все доски

```http
GET /api/v1/boards
```

**Ответ:**

```json
{
  "success": true,
  "data": [
    {
      "id": "b",
      "name": "Random",
      "description": "Random topics",
      "thread_count": 5,
      "created_at": "2025-12-06T10:00:00Z"
    },
    {
      "id": "pr",
      "name": "Программирование",
      "description": "Pair of programming",
      "thread_count": 12,
      "created_at": "2025-12-06T11:30:00Z"
    }
  ]
}
```

### Получить доску

```http
GET /api/v1/boards/{id}
```

**Параметры:**
- `id` — ID доски

**Пример:** `GET /api/v1/boards/b`

**Ответ:**

```json
{
  "success": true,
  "data": {
    "id": "b",
    "name": "Random",
    "description": "Random topics",
    "created_at": "2025-12-06T10:00:00Z"
  }
}
```

**Ошибки:**
- `404` — Доска не найдена

### Создать доску

```http
POST /api/v1/boards
Content-Type: application/json
```

**Тело запроса:**

```json
{
  "id": "tech",
  "name": "Технологии",
  "description": "Pair of technology discussions"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| id | string | ✅ | Только a-z и 0-9 |
| name | string | ✅ | Название |
| description | string | ❌ | Описание |

**Ответ:**

```json
{
  "success": true,
  "data": {
    "id": "tech",
    "message": "Доска создана"
  }
}
```

**Ошибки:**
- `400` — Неверные данные
- `409` — Доска уже существует

---

## Треды

### Получить треды доски

```http
GET /api/v1/boards/{id}/threads?sort=bump
```

**Параметры:**
- `id` — ID доски
- `sort` — Сортировка: `bump` (по умолчанию), `new`, `old`, `replies`

**Ответ:**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "board_id": "b",
      "subject": "Тема треда",
      "post_count": 15,
      "created_at": "2025-12-06T10:00:00Z",
      "bumped_at": "2025-12-06T14:30:00Z",
      "first_post": {
        "id": 1,
        "author": "Аноним",
        "content": "Текст первого поста...",
        "media_path": "/uploads/1_123.jpg",
        "media_type": "image",
        "created_at": "2025-12-06T10:00:00Z"
      }
    }
  ]
}
```

### Получить тред с постами

```http
GET /api/v1/threads/{id}
```

**Ответ:**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "board_id": "b",
    "subject": "Тема треда",
    "post_count": 3,
    "created_at": "2025-12-06T10:00:00Z",
    "bumped_at": "2025-12-06T14:30:00Z",
    "posts": [
      {
        "id": 1,
        "thread_id": 1,
        "parent_id": 0,
        "author": "Аноним",
        "content": "Первый пост",
        "media_path": "/uploads/1_123.jpg",
        "media_type": "image",
        "created_at": "2025-12-06T10:00:00Z",
        "depth": 0
      },
      {
        "id": 2,
        "thread_id": 1,
        "parent_id": 1,
        "author": "Аноним",
        "content": "Ответ на первый",
        "created_at": "2025-12-06T10:05:00Z",
        "depth": 1
      }
    ]
  }
}
```

### Создать тред

```http
POST /api/v1/threads
Content-Type: application/json
```

**Тело запроса:**

```json
{
  "board_id": "b",
  "subject": "Новый тред",
  "author": "Аноним",
  "content": "Текст первого поста",
  "media_path": "/uploads/123.jpg",
  "media_type": "image"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| board_id | string | ✅ | ID доски |
| subject | string | ✅ | Тема треда |
| content | string | ✅ | Текст первого поста |
| author | string | ❌ | По умолчанию "Аноним" |
| media_path | string | ❌ | Путь от /api/v1/upload |
| media_type | string | ❌ | image/video/audio |

**Ответ:**

```json
{
  "success": true,
  "data": {
    "thread_id": 5,
    "post_id": 10,
    "message": "Тред создан"
  }
}
```

---

## Посты

### Создать пост

```http
POST /api/v1/posts
Content-Type: application/json
```

**Тело запроса:**

```json
{
  "thread_id": 1,
  "parent_id": 0,
  "author": "Аноним",
  "content": "Текст ответа",
  "media_path": "/uploads/456.mp4",
  "media_type": "video"
}
```

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| thread_id | int | ✅ | ID треда |
| content | string | ✅ | Текст поста |
| parent_id | int | ❌ | ID родительского поста |
| author | string | ❌ | По умолчанию "Аноним" |
| media_path | string | ❌ | Путь от /api/v1/upload |
| media_type | string | ❌ | image/video/audio |

**Ответ:**

```json
{
  "success": true,
  "data": {
    "post_id": 15,
    "message": "Пост создан"
  }
}
```

---

## Загрузка файлов

### Загрузить медиафайл

```http
POST /api/v1/upload
Content-Type: multipart/form-data
```

**Параметры формы:**
- `media` — файл (обязательно)

**Пример cURL:**

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "media=@image.jpg"
```

**Ответ:**

```json
{
  "success": true,
  "data": {
    "path": "/uploads/1733500800_123456789.jpg",
    "type": "image"
  }
}
```

**Поддерживаемые форматы:**

| Тип | Расширения |
|-----|------------|
| image | jpg, jpeg, png, gif, webp, svg |
| video | mp4, webm, avi, mov, mkv |
| audio | mp3, wav, ogg, flac, m4a |

**Лимит:** 100MB

### Использование с постом

1. Загрузите файл через `/api/v1/upload`
2. Получите `path` и `type` из ответа
3. Используйте в `/api/v1/posts`:

```json
{
  "thread_id": 1,
  "content": "Пост с картинкой",
  "media_path": "/uploads/1733500800_123456789.jpg",
  "media_type": "image"
}
```

---

## Примеры cURL

### Получить доски

```bash
curl http://localhost:8080/api/v1/boards
```

### Создать доску

```bash
curl -X POST http://localhost:8080/api/v1/boards \
  -H "Content-Type: application/json" \
  -d '{"id":"test","name":"Тест","description":"Тестовая доска"}'
```

### Получить треды

```bash
curl http://localhost:8080/api/v1/boards/b/threads?sort=new
```

### Создать тред

```bash
curl -X POST http://localhost:8080/api/v1/threads \
  -H "Content-Type: application/json" \
  -d '{"board_id":"b","subject":"Новый тред","content":"Привет!"}'
```

### Создать пост

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{"thread_id":1,"content":"Ответ","parent_id":1}'
```

---

## Коды ошибок

| Код | Описание |
|-----|----------|
| 200 | Успешно |
| 400 | Неверные данные запроса |
| 404 | Ресурс не найден |
| 409 | Конфликт (уже существует) |
| 500 | Внутренняя ошибка сервера |

