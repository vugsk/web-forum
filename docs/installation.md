# Установка и запуск

## Требования

- **Go** 1.21 или выше — [golang.org/dl](https://golang.org/dl/)
- **MySQL** 8.0 или выше — [mysql.com/downloads](https://www.mysql.com/downloads/)
- **Git** (опционально) — [git-scm.com](https://git-scm.com/)

## Установка

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-username/webForum.git
cd webForum
```

Или скачайте и распакуйте архив.

### 2. Установка зависимостей Go

```bash
go mod download
```

Это установит:
- `github.com/go-sql-driver/mysql` — драйвер MySQL
- `github.com/gorilla/websocket` — WebSocket
- `github.com/joho/godotenv` — загрузка .env файлов

### 3. Создание базы данных

Подключитесь к MySQL и выполните:

```sql
CREATE DATABASE webforum 
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;
```

Таблицы создадутся автоматически при первом запуске.

### 4. Настройка конфигурации

Создайте файл `.env` в корне проекта:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
# База данных
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=ваш_пароль
DB_NAME=webforum

# Порт сервера
PORT=8080
```

### 5. Запуск

#### Режим разработки

```bash
go run main.go
```

#### Сборка и запуск

```bash
# Windows
go build -o forum.exe .
.\forum.exe

# Linux/macOS
go build -o forum .
./forum
```

### 6. Проверка

Откройте в браузере: http://localhost:8080

Вы должны увидеть главную страницу форума.

## Возможные проблемы

### Ошибка подключения к БД

```
Ошибка подключения к БД: dial tcp 127.0.0.1:3306: connect: connection refused
```

**Решение:** Убедитесь, что MySQL запущен и данные в `.env` верны.

### Порт уже занят

```
listen tcp :8080: bind: address already in use
```

**Решение:** Измените порт в `.env` или завершите процесс на порту 8080:

```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/macOS
lsof -i :8080
kill -9 <PID>
```

### Ошибка шаблонов

```
panic: template: pattern matches no files
```

**Решение:** Убедитесь, что папка `templates/` существует и содержит HTML файлы.

## Структура после установки

```
webForum/
├── main.go
├── go.mod
├── go.sum
├── .env              # ваша конфигурация
├── .env.example
├── forum.exe         # скомпилированный файл
├── database/
├── handlers/
├── static/
├── templates/
├── uploads/          # создаётся автоматически
└── docs/
```

