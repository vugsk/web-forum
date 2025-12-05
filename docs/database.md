# База данных

## Обзор

Проект использует MySQL 8.0+ с кодировкой `utf8mb4` для полной поддержки Unicode (включая эмодзи).

## Схема

### Таблица `boards` (Доски)

```sql
CREATE TABLE boards (
    id VARCHAR(50) PRIMARY KEY,           -- ID доски (например: "b", "pr")
    name VARCHAR(255) NOT NULL,           -- Название
    description TEXT,                      -- Описание
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | VARCHAR(50) | Уникальный ID (латиница + цифры) |
| `name` | VARCHAR(255) | Отображаемое название |
| `description` | TEXT | Описание доски |
| `created_at` | TIMESTAMP | Дата создания |

### Таблица `threads` (Треды)

```sql
CREATE TABLE threads (
    id INT AUTO_INCREMENT PRIMARY KEY,
    board_id VARCHAR(50) NOT NULL,         -- Ссылка на доску
    subject VARCHAR(255) NOT NULL,         -- Тема треда
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    bumped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    INDEX idx_board_bumped (board_id, bumped_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INT | Автоинкрементный ID |
| `board_id` | VARCHAR(50) | FK на boards.id |
| `subject` | VARCHAR(255) | Тема/заголовок |
| `created_at` | TIMESTAMP | Дата создания |
| `bumped_at` | TIMESTAMP | Время последнего бампа |

### Таблица `posts` (Посты)

```sql
CREATE TABLE posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    thread_id INT NOT NULL,                -- Ссылка на тред
    parent_id INT DEFAULT NULL,            -- Ссылка на родительский пост
    author VARCHAR(100) DEFAULT 'Аноним',  -- Имя автора
    content TEXT NOT NULL,                 -- Текст поста
    media_path VARCHAR(500),               -- Путь к файлу
    media_type VARCHAR(20),                -- Тип: image/video/audio
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES posts(id) ON DELETE SET NULL,
    INDEX idx_thread (thread_id),
    INDEX idx_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INT | Автоинкрементный ID |
| `thread_id` | INT | FK на threads.id |
| `parent_id` | INT | FK на posts.id (для ответов) |
| `author` | VARCHAR(100) | Имя автора |
| `content` | TEXT | Текст сообщения |
| `media_path` | VARCHAR(500) | Путь к медиафайлу |
| `media_type` | VARCHAR(20) | Тип медиа |
| `created_at` | TIMESTAMP | Дата создания |

## Связи

```
boards (1) ──────< threads (N)
                      │
                      │
threads (1) ─────< posts (N)
                      │
                      │ parent_id (self-reference)
                      │
posts (1) ───────< posts (N)
```

## Индексы

| Таблица | Индекс | Назначение |
|---------|--------|------------|
| threads | `idx_board_bumped` | Быстрая сортировка по бампу |
| posts | `idx_thread` | Быстрый поиск постов треда |
| posts | `idx_parent` | Построение дерева ответов |

## Каскадное удаление

- При удалении **доски** → удаляются все её **треды**
- При удалении **треда** → удаляются все его **посты**
- При удалении **поста** → у дочерних постов `parent_id = NULL`

## Примеры запросов

### Получить все доски с количеством тредов

```sql
SELECT b.id, b.name, b.description, b.created_at,
       COUNT(t.id) as thread_count
FROM boards b
LEFT JOIN threads t ON b.id = t.board_id
GROUP BY b.id
ORDER BY b.id;
```

### Получить треды доски, отсортированные по бампу

```sql
SELECT t.id, t.board_id, t.subject, t.created_at, t.bumped_at,
       COUNT(p.id) as post_count
FROM threads t
LEFT JOIN posts p ON t.id = p.thread_id
WHERE t.board_id = 'b'
GROUP BY t.id
ORDER BY t.bumped_at DESC;
```

### Получить посты треда

```sql
SELECT id, thread_id, parent_id, author, content,
       media_path, media_type, created_at
FROM posts
WHERE thread_id = 1
ORDER BY created_at ASC;
```

### Бамп треда

```sql
UPDATE threads 
SET bumped_at = CURRENT_TIMESTAMP 
WHERE id = 1;
```

## Инициализация

Таблицы создаются автоматически при запуске через `database.InitSchema()`.

Для ручного создания используйте файл `database/schema.sql`:

```bash
mysql -u root -p webforum < database/schema.sql
```

## Резервное копирование

```bash
# Бэкап
mysqldump -u root -p webforum > backup.sql

# Восстановление
mysql -u root -p webforum < backup.sql
```

