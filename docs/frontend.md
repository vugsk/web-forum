# Фронтенд

## Обзор

Фронтенд реализован на чистом HTML, CSS и JavaScript без фреймворков.

## Шаблоны (templates/)

### Система шаблонов

Используется стандартный Go пакет `html/template`.

**Функции шаблонов:**

| Функция | Описание | Пример |
|---------|----------|--------|
| `formatTime` | Форматирование даты | `{{formatTime .CreatedAt}}` → "06.12.2025 14:30:00" |
| `truncate` | Обрезка текста | `{{truncate .Content 300}}` |
| `multiply` | Умножение (для отступов) | `{{multiply .Depth 20}}px` |
| `nullStr` | sql.NullString → string | `{{nullStr .MediaPath}}` |
| `nullInt` | sql.NullInt64 → int | `{{nullInt .ParentID}}` |

### index.html (Главная страница)

```html
<!-- Список досок -->
{{range .Boards}}
<div class="board-item" id="board-{{.ID}}">
    <h3><a href="/board/{{.ID}}">/{{.ID}}/ - {{.Name}}</a></h3>
    <p>{{.Description}}</p>
    <p class="stats">Тредов: {{.ThreadCount}}</p>
</div>
{{else}}
<p class="no-content">Пока нет досок.</p>
{{end}}
```

**Особенности:**
- Поиск досок (JavaScript)
- Модальное окно создания доски
- WebSocket для live-обновлений

### board.html (Страница доски)

```html
<!-- Форма создания треда -->
<form action="/api/thread" method="POST" enctype="multipart/form-data">
    <input type="hidden" name="board_id" value="{{.Board.ID}}">
    <input type="text" name="subject" required>
    <textarea name="content" required></textarea>
    <input type="file" name="media">
</form>

<!-- Список тредов -->
{{range .Threads}}
<div class="thread-preview" id="thread-{{.ID}}">
    <strong>{{.Subject}}</strong>
    <!-- ... -->
</div>
{{end}}
```

**Особенности:**
- Сортировка тредов
- Превью первого поста
- WebSocket для новых тредов

### thread.html (Страница треда)

```html
<!-- Список постов -->
{{range $index, $post := .Posts}}
<div class="post" id="post-{{$post.ID}}" 
     style="margin-left: {{multiply $post.Depth 20}}px;">
    <span class="post-author">{{$post.Author}}</span>
    {{if $post.ParentID.Valid}}
    <span class="reply-to">&gt;&gt;{{nullInt $post.ParentID}}</span>
    {{end}}
    <p>{{$post.Content}}</p>
</div>
{{end}}
```

**Особенности:**
- Древовидная структура (отступы по глубине)
- Кнопка "Ответить" на каждом посте
- WebSocket для новых постов

## Стили (static/style.css)

### Цветовая схема (4chan-like)

```css
:root {
    --bg-page: #eef2ff;        /* Фон страницы */
    --bg-block: #d6daf0;       /* Фон блоков */
    --bg-post: #f0e0d6;        /* Фон постов */
    --text-title: #af0a0f;     /* Заголовки */
    --text-author: #117743;    /* Имя автора */
    --text-link: #34345c;      /* Ссылки */
    --text-green: #789922;     /* Greentext */
}
```

### Основные классы

| Класс | Описание |
|-------|----------|
| `.container` | Контейнер страницы (max-width: 1000px) |
| `.board-item` | Карточка доски |
| `.thread-preview` | Превью треда |
| `.post` | Пост в треде |
| `.op-post` | Первый пост (OP) |
| `.new-post` | Анимация нового поста |

### Анимации

```css
/* Новый пост */
.new-post {
    animation: newPostHighlight 2s ease-out;
}

@keyframes newPostHighlight {
    0% { background-color: #c8ffc8; }
    100% { background-color: #f0e0d6; }
}

/* Обновление треда */
.thread-updated {
    animation: threadUpdatedHighlight 2s ease-out;
}

@keyframes threadUpdatedHighlight {
    0% { background-color: #ffffc8; }
    100% { background-color: #d6daf0; }
}
```

### Модальное окно

```css
.modal-overlay {
    display: none;
    position: fixed;
    top: 0; left: 0;
    width: 100%; height: 100%;
    background: rgba(0,0,0,0.5);
}

.modal-overlay.active {
    display: flex;
    justify-content: center;
    align-items: center;
}

.modal {
    background: #d6daf0;
    padding: 20px;
    max-width: 450px;
}
```

### Адаптивность

```css
@media (max-width: 768px) {
    .container { padding: 5px; }
    .post { margin-left: 0 !important; }
    .post-media { float: none; }
}
```

## JavaScript

### WebSocket подключение

```javascript
function connectWebSocket() {
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(protocol + '//' + location.host + '/ws/thread?thread_id=' + threadID);
    
    ws.onopen = () => {
        document.getElementById('ws-status').textContent = '🟢 Live';
    };
    
    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        if (msg.type === 'new_post') {
            addNewPost(msg.data);
        }
    };
    
    ws.onclose = () => {
        document.getElementById('ws-status').textContent = '🔴 Offline';
        setTimeout(connectWebSocket, 3000);
    };
}
```

### Добавление поста

```javascript
function addNewPost(postData) {
    // Проверка дубликата
    if (document.getElementById('post-' + postData.id)) return;
    
    // Вычисление отступа
    let depth = 0;
    if (postData.parent_id > 0) {
        const parent = document.getElementById('post-' + postData.parent_id);
        depth = parseInt(parent?.dataset.depth || 0) + 1;
    }
    
    // Создание HTML
    const html = `<div class="post new-post" id="post-${postData.id}" 
                       style="margin-left: ${depth * 20}px">...</div>`;
    
    document.getElementById('thread-posts').insertAdjacentHTML('beforeend', html);
    
    // Убираем анимацию через 2 сек
    setTimeout(() => {
        document.getElementById('post-' + postData.id).classList.remove('new-post');
    }, 2000);
}
```

### Модальное окно

```javascript
function openModal() {
    document.getElementById('modal-overlay').classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeModal(event) {
    if (event && event.target !== event.currentTarget) return;
    document.getElementById('modal-overlay').classList.remove('active');
    document.body.style.overflow = '';
}

// Закрытие по Escape
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
});
```

### Поиск досок

```javascript
document.getElementById('board-search').addEventListener('input', function() {
    const query = this.value.toLowerCase();
    
    document.querySelectorAll('.board-item').forEach(board => {
        const name = board.dataset.name.toLowerCase();
        const id = board.dataset.id.toLowerCase();
        const match = name.includes(query) || id.includes(query);
        board.style.display = match ? 'block' : 'none';
    });
});
```

### Ответ на пост

```javascript
function setReplyTo(postId) {
    document.getElementById('parent_id').value = postId;
    document.getElementById('reply-to-id').textContent = '>>' + postId;
    document.getElementById('reply-info').style.display = 'flex';
    document.getElementById('reply-form').scrollIntoView({ behavior: 'smooth' });
}

function clearReply() {
    document.getElementById('parent_id').value = '0';
    document.getElementById('reply-info').style.display = 'none';
}
```

## Загрузка файлов

### HTML форма

```html
<form action="/api/thread" method="POST" enctype="multipart/form-data">
    <input type="file" name="media" accept="image/*,video/*,audio/*">
    <span class="file-hint">Макс. 100MB</span>
</form>
```

### Отображение медиа

```html
{{if eq (nullStr .MediaType) "image"}}
<img src="{{nullStr .MediaPath}}" class="media-image">

{{else if eq (nullStr .MediaType) "video"}}
<video controls class="media-video">
    <source src="{{nullStr .MediaPath}}">
</video>

{{else if eq (nullStr .MediaType) "audio"}}
<audio controls class="media-audio">
    <source src="{{nullStr .MediaPath}}">
</audio>
{{end}}
```

