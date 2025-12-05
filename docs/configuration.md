# Конфигурация

## Переменные окружения

Конфигурация загружается из файла `.env` или переменных окружения системы.

### Создание файла .env

```bash
cp .env.example .env
```

### Параметры

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост MySQL | `localhost` |
| `DB_PORT` | Порт MySQL | `3306` |
| `DB_USER` | Пользователь MySQL | `root` |
| `DB_PASSWORD` | Пароль MySQL | `` (пустой) |
| `DB_NAME` | Имя базы данных | `webforum` |
| `PORT` | Порт HTTP сервера | `8080` |

### Пример .env

```env
# База данных MySQL
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=mypassword123
DB_NAME=webforum

# Сервер
PORT=8080
```

## Приоритет конфигурации

1. Переменные окружения системы
2. Файл `.env`
3. Значения по умолчанию в коде

## Загрузка конфигурации

```go
// main.go
func main() {
    // Загрузка .env (если есть)
    if err := godotenv.Load(); err != nil {
        log.Println("Файл .env не найден")
    }
    
    // Использование с fallback
    dbConfig := database.Config{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     getEnv("DB_PORT", "3306"),
        User:     getEnv("DB_USER", "root"),
        Password: getEnv("DB_PASSWORD", ""),
        DBName:   getEnv("DB_NAME", "webforum"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Конфигурация для разных окружений

### Разработка (development)

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=webforum_dev
PORT=8080
```

### Продакшен (production)

```env
DB_HOST=mysql.example.com
DB_PORT=3306
DB_USER=forum_user
DB_PASSWORD=strong_password_here
DB_NAME=webforum_prod
PORT=80
```

### Docker

```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=forum
DB_PASSWORD=forum_password
DB_NAME=webforum
PORT=8080
```

## Параметры MySQL

### Строка подключения

```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
    cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
```

### Опции DSN

| Параметр | Описание |
|----------|----------|
| `charset=utf8mb4` | Полная поддержка Unicode |
| `parseTime=True` | Парсинг DATETIME в time.Time |
| `loc=Local` | Локальная временная зона |

## Лимиты

### Размер файлов

```go
// handlers/handlers.go
r.ParseMultipartForm(100 << 20) // 100MB
```

### Типы файлов

```go
var allowedExtensions = map[string]string{
    ".jpg": "image", ".jpeg": "image", ".png": "image",
    ".gif": "image", ".webp": "image", ".svg": "image",
    ".mp4": "video", ".webm": "video", ".avi": "video",
    ".mov": "video", ".mkv": "video",
    ".mp3": "audio", ".wav": "audio", ".ogg": "audio",
    ".flac": "audio", ".m4a": "audio",
}
```

## WebSocket настройки

```go
// handlers/websocket.go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Разрешить все origins
    },
}
```

**Для продакшена** рекомендуется ограничить origins:

```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "https://yourdomain.com"
}
```

## Безопасность

### В .gitignore

```gitignore
.env
.env.local
.env.*.local
```

### Рекомендации

1. Никогда не коммитьте `.env` с паролями
2. Используйте сложные пароли в продакшене
3. Ограничьте права пользователя MySQL
4. Используйте HTTPS в продакшене

