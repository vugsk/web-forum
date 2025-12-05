# Архитектура проекта

## Структура директорий

```
webForum/
├── main.go                 # Точка входа, маршрутизация
├── go.mod                  # Go модуль
├── go.sum                  # Контрольные суммы зависимостей
├── .env                    # Конфигурация (не в git)
├── .env.example            # Пример конфигурации
├── .gitignore              # Игнорируемые файлы
├── LICENSE                 # MIT лицензия
├── README.md               # Описание проекта
│
├── database/               # Слой работы с БД
│   ├── database.go         # Подключение, настройка пула
│   ├── queries.go          # SQL-запросы, CRUD операции
│   └── schema.sql          # SQL-схема для ручного создания
│
├── handlers/               # HTTP обработчики
│   ├── handlers.go         # Веб-страницы и формы
│   ├── api.go              # REST API v1
│   └── websocket.go        # WebSocket хаб и обработчики
│
├── static/                 # Статические файлы
│   └── style.css           # Стили (4chan-like)
│
├── templates/              # HTML шаблоны
│   ├── index.html          # Главная страница
│   ├── board.html          # Страница доски
│   └── thread.html         # Страница треда
│
├── uploads/                # Загруженные файлы (не в git)
│   └── ...
│
└── docs/                   # Документация
    └── ...
```

## Компоненты

### main.go

Точка входа приложения:
- Загрузка конфигурации из `.env`
- Подключение к MySQL
- Инициализация таблиц
- Настройка маршрутов
- Запуск HTTP сервера

```go
func main() {
    // 1. Загрузка .env
    godotenv.Load()
    
    // 2. Подключение к БД
    database.Connect(config)
    database.InitSchema()
    
    // 3. Маршруты
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler.IndexHandler)
    // ...
    
    // 4. Запуск
    http.ListenAndServe(":8080", mux)
}
```

### database/

#### database.go
- `Connect(cfg Config)` — подключение к MySQL
- `Close()` — закрытие соединения
- `InitSchema()` — создание таблиц

#### queries.go
- `GetAllBoards()` — все доски
- `GetBoard(id)` — доска по ID
- `CreateBoard(id, name, desc)` — создание доски
- `GetThreadsByBoard(boardID, sort)` — треды доски
- `GetThread(id)` — тред по ID
- `CreateThread(boardID, subject)` — создание треда
- `BumpThread(id)` — обновление времени бампа
- `GetPostsByThread(threadID)` — посты треда
- `CreatePost(...)` — создание поста

### handlers/

#### handlers.go
Веб-обработчики для HTML страниц:
- `IndexHandler` — главная (`/`)
- `BoardHandler` — доска (`/board/{id}`)
- `ThreadHandler` — тред (`/thread/{id}`)
- `CreateBoardHandler` — создание доски
- `CreateThreadHandler` — создание треда
- `CreatePostHandler` — создание поста

#### api.go
REST API для мобильных приложений:
- `APIGetBoards` — GET/POST `/api/v1/boards`
- `APIGetBoard` — GET `/api/v1/boards/{id}`
- `APIGetThreads` — GET `/api/v1/boards/{id}/threads`
- `APIGetThread` — GET `/api/v1/threads/{id}`
- `APICreateThread` — POST `/api/v1/threads`
- `APICreatePost` — POST `/api/v1/posts`
- `APIUploadMedia` — POST `/api/v1/upload`

#### websocket.go
WebSocket для live-обновлений:
- `Hub` — управление соединениями
- `WebSocketHomeHandler` — главная страница
- `WebSocketBoardHandler` — страница доски
- `WebSocketThreadHandler` — страница треда

## Поток данных

### Создание поста

```
Browser                    Server                     Database
   │                          │                          │
   │  POST /api/post          │                          │
   │ ─────────────────────>   │                          │
   │                          │  INSERT INTO posts       │
   │                          │ ─────────────────────>   │
   │                          │                          │
   │                          │  UPDATE threads (bump)   │
   │                          │ ─────────────────────>   │
   │                          │                          │
   │                          │  WebSocket broadcast     │
   │  <─────────────────────  │  to thread clients       │
   │  (new post appears)      │                          │
   │                          │                          │
   │  Redirect to thread      │                          │
   │ <─────────────────────   │                          │
```

### WebSocket подключение

```
Browser                    Server
   │                          │
   │  WS /ws/thread?id=1      │
   │ ─────────────────────>   │
   │                          │
   │  Connection established  │
   │ <─────────────────────   │
   │                          │
   │  Register in Hub         │
   │  (threadClients[1])      │
   │                          │
   │      ... waiting ...     │
   │                          │
   │  {"type":"new_post",...} │
   │ <─────────────────────   │
   │                          │
   │  JavaScript adds post    │
   │  to DOM                  │
```

## Зависимости

```go
require (
    github.com/go-sql-driver/mysql v1.8.1  // MySQL драйвер
    github.com/gorilla/websocket v1.5.3    // WebSocket
    github.com/joho/godotenv v1.5.1        // .env файлы
)
```

