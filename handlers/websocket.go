package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все origins для разработки
	},
}

// Сообщение для отправки клиентам
type WSMessage struct {
	Type     string      `json:"type"` // "new_post", "new_thread", "new_board"
	ThreadID int         `json:"thread_id,omitempty"`
	BoardID  string      `json:"board_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

// Hub управляет всеми WebSocket соединениями
type Hub struct {
	// Клиенты по тредам: thread_id -> []*websocket.Conn
	threadClients map[int]map[*websocket.Conn]bool
	// Клиенты по доскам: board_id -> []*websocket.Conn
	boardClients map[string]map[*websocket.Conn]bool
	// Мьютекс для безопасного доступа
	mu sync.RWMutex
}

// Глобальный хаб
var WsHub = &Hub{
	threadClients: make(map[int]map[*websocket.Conn]bool),
	boardClients:  make(map[string]map[*websocket.Conn]bool),
}

// RegisterThreadClient регистрирует клиента для треда
func (h *Hub) RegisterThreadClient(threadID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.threadClients[threadID] == nil {
		h.threadClients[threadID] = make(map[*websocket.Conn]bool)
	}
	h.threadClients[threadID][conn] = true
	log.Printf("WebSocket: клиент подключился к треду #%d (всего: %d)", threadID, len(h.threadClients[threadID]))
}

// UnregisterThreadClient удаляет клиента из треда
func (h *Hub) UnregisterThreadClient(threadID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.threadClients[threadID] != nil {
		delete(h.threadClients[threadID], conn)
		log.Printf("WebSocket: клиент отключился от треда #%d (осталось: %d)", threadID, len(h.threadClients[threadID]))
	}
}

// RegisterBoardClient регистрирует клиента для доски
func (h *Hub) RegisterBoardClient(boardID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.boardClients[boardID] == nil {
		h.boardClients[boardID] = make(map[*websocket.Conn]bool)
	}
	h.boardClients[boardID][conn] = true
	log.Printf("WebSocket: клиент подключился к доске /%s/ (всего: %d)", boardID, len(h.boardClients[boardID]))
}

// UnregisterBoardClient удаляет клиента из доски
func (h *Hub) UnregisterBoardClient(boardID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.boardClients[boardID] != nil {
		delete(h.boardClients[boardID], conn)
		log.Printf("WebSocket: клиент отключился от доски /%s/ (осталось: %d)", boardID, len(h.boardClients[boardID]))
	}
}

// BroadcastToThread отправляет сообщение всем клиентам треда
func (h *Hub) BroadcastToThread(threadID int, msg WSMessage) {
	h.mu.RLock()
	clients := h.threadClients[threadID]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WebSocket: ошибка сериализации: %v", err)
		return
	}

	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("WebSocket: ошибка отправки: %v", err)
			conn.Close()
			h.UnregisterThreadClient(threadID, conn)
		}
	}
}

// BroadcastToBoard отправляет сообщение всем клиентам доски
func (h *Hub) BroadcastToBoard(boardID string, msg WSMessage) {
	h.mu.RLock()
	clients := h.boardClients[boardID]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("WebSocket: ошибка сериализации: %v", err)
		return
	}

	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("WebSocket: ошибка отправки: %v", err)
			conn.Close()
			h.UnregisterBoardClient(boardID, conn)
		}
	}
}

// WebSocketThreadHandler обрабатывает WebSocket соединения для треда
func WebSocketThreadHandler(w http.ResponseWriter, r *http.Request) {
	threadIDStr := r.URL.Query().Get("thread_id")
	if threadIDStr == "" {
		http.Error(w, "thread_id required", http.StatusBadRequest)
		return
	}

	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "invalid thread_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	WsHub.RegisterThreadClient(threadID, conn)

	// Читаем сообщения (для поддержания соединения)
	go func() {
		defer func() {
			WsHub.UnregisterThreadClient(threadID, conn)
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// WebSocketBoardHandler обрабатывает WebSocket соединения для доски
func WebSocketBoardHandler(w http.ResponseWriter, r *http.Request) {
	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		http.Error(w, "board_id required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	WsHub.RegisterBoardClient(boardID, conn)

	// Читаем сообщения (для поддержания соединения)
	go func() {
		defer func() {
			WsHub.UnregisterBoardClient(boardID, conn)
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
