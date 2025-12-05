# WebSocket API

## Обзор

WebSocket используется для live-обновлений: новые посты, треды и доски появляются мгновенно без перезагрузки страницы.

## Endpoints

| URL | Описание |
|-----|----------|
| `/ws/home` | Обновления главной страницы |
| `/ws/board?board_id={id}` | Обновления доски |
| `/ws/thread?thread_id={id}` | Обновления треда |

## Подключение

### JavaScript

```javascript
// Определяем протокол (ws или wss)
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

// Подключение к треду
const ws = new WebSocket(protocol + '//' + window.location.host + '/ws/thread?thread_id=1');

ws.onopen = function() {
    console.log('Подключено');
};

ws.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    console.log('Получено:', msg);
};

ws.onclose = function() {
    console.log('Отключено');
    // Переподключение через 3 секунды
    setTimeout(connectWebSocket, 3000);
};

ws.onerror = function(error) {
    console.error('Ошибка:', error);
};
```

### Параметры подключения

| Endpoint | Параметр | Описание |
|----------|----------|----------|
| `/ws/home` | — | Без параметров |
| `/ws/board` | `board_id` | ID доски |
| `/ws/thread` | `thread_id` | ID треда |

## Формат сообщений

### Базовая структура

```json
{
  "type": "тип_события",
  "thread_id": 123,
  "board_id": "b",
  "data": { ... }
}
```

## Типы событий

### `new_board`

Отправляется на `/ws/home` при создании новой доски.

```json
{
  "type": "new_board",
  "data": {
    "id": "tech",
    "name": "Технологии",
    "description": "Pair of technology"
  }
}
```

### `new_thread`

Отправляется на `/ws/board` при создании нового треда.

```json
{
  "type": "new_thread",
  "thread_id": 5,
  "board_id": "b",
  "data": {
    "id": 5,
    "post_id": 10,
    "subject": "Новый тред",
    "author": "Аноним",
    "content": "Текст первого поста",
    "media_path": "/uploads/123.jpg",
    "media_type": "image",
    "created_at": "06.12.2025 14:30:00"
  }
}
```

### `thread_updated`

Отправляется на `/ws/board` при новом посте в треде.

```json
{
  "type": "thread_updated",
  "thread_id": 5,
  "board_id": "b"
}
```

### `new_post`

Отправляется на `/ws/thread` при создании нового поста.

```json
{
  "type": "new_post",
  "thread_id": 1,
  "data": {
    "id": 15,
    "author": "Аноним",
    "content": "Текст ответа",
    "media_path": "/uploads/456.mp3",
    "media_type": "audio",
    "parent_id": 10,
    "created_at": "06.12.2025 14:35:00"
  }
}
```

## Обработка на клиенте

### Новый пост в треде

```javascript
ws.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    
    if (msg.type === 'new_post') {
        const post = msg.data;
        
        // Проверяем, нет ли уже поста
        if (document.getElementById('post-' + post.id)) {
            return;
        }
        
        // Создаём HTML
        const html = `
            <div class="post" id="post-${post.id}">
                <span class="post-author">${post.author}</span>
                <span class="post-date">${post.created_at}</span>
                <p>${post.content}</p>
            </div>
        `;
        
        // Добавляем в DOM
        document.getElementById('posts').insertAdjacentHTML('beforeend', html);
    }
};
```

### Новый тред на доске

```javascript
ws.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    
    if (msg.type === 'new_thread') {
        // Добавить тред в начало списка
        addNewThread(msg.data);
    } else if (msg.type === 'thread_updated') {
        // Обновить счётчик постов
        updateThreadCounter(msg.thread_id);
    }
};
```

### Новая доска на главной

```javascript
ws.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    
    if (msg.type === 'new_board') {
        addNewBoard(msg.data);
    }
};
```

## Переподключение

Рекомендуется реализовать автоматическое переподключение:

```javascript
let ws;
let reconnectInterval;

function connect() {
    ws = new WebSocket(wsUrl);
    
    ws.onopen = function() {
        console.log('Подключено');
        clearInterval(reconnectInterval);
        reconnectInterval = null;
    };
    
    ws.onclose = function() {
        console.log('Отключено');
        
        // Переподключение каждые 3 секунды
        if (!reconnectInterval) {
            reconnectInterval = setInterval(connect, 3000);
        }
    };
}

connect();
```

## Индикатор статуса

В шаблонах реализован индикатор:
- 🟢 **Live** — подключено
- 🔴 **Offline** — отключено

```javascript
ws.onopen = function() {
    document.getElementById('ws-status').textContent = '🟢 Live';
    document.getElementById('ws-status').style.color = '#117743';
};

ws.onclose = function() {
    document.getElementById('ws-status').textContent = '🔴 Offline';
    document.getElementById('ws-status').style.color = '#af0a0f';
};
```

## Серверная часть

### Hub (handlers/websocket.go)

```go
type Hub struct {
    threadClients map[int]map[*websocket.Conn]bool
    boardClients  map[string]map[*websocket.Conn]bool
    homeClients   map[*websocket.Conn]bool
    mu            sync.RWMutex
}

// Отправка сообщения в тред
func (h *Hub) BroadcastToThread(threadID int, msg WSMessage) {
    // ...
}

// Отправка сообщения на доску
func (h *Hub) BroadcastToBoard(boardID string, msg WSMessage) {
    // ...
}

// Отправка на главную
func (h *Hub) BroadcastToHome(msg WSMessage) {
    // ...
}
```

### Вызов из обработчиков

```go
// При создании поста
WsHub.BroadcastToThread(threadID, WSMessage{
    Type:     "new_post",
    ThreadID: threadID,
    Data: map[string]interface{}{
        "id":      postID,
        "author":  author,
        "content": content,
        // ...
    },
})

// Также уведомляем доску
WsHub.BroadcastToBoard(boardID, WSMessage{
    Type:     "thread_updated",
    ThreadID: threadID,
    BoardID:  boardID,
})
```

## Безопасность

- `CheckOrigin` в upgrader разрешает все origins (для разработки)
- В продакшене рекомендуется ограничить origins
- Соединения автоматически закрываются при ошибках

