package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"webForum/database"
)

// APIResponse стандартный ответ API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BoardResponse доска для API
type BoardResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ThreadCount int    `json:"thread_count"`
	CreatedAt   string `json:"created_at"`
}

// ThreadResponse тред для API
type ThreadResponse struct {
	ID        int            `json:"id"`
	BoardID   string         `json:"board_id"`
	Subject   string         `json:"subject"`
	PostCount int            `json:"post_count"`
	CreatedAt string         `json:"created_at"`
	BumpedAt  string         `json:"bumped_at"`
	FirstPost *PostResponse  `json:"first_post,omitempty"`
	Posts     []PostResponse `json:"posts,omitempty"`
}

// PostResponse пост для API
type PostResponse struct {
	ID        int    `json:"id"`
	ThreadID  int    `json:"thread_id"`
	ParentID  int    `json:"parent_id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	MediaPath string `json:"media_path,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	CreatedAt string `json:"created_at"`
	Depth     int    `json:"depth"`
}

// Helper для отправки JSON
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendSuccess(w http.ResponseWriter, data interface{}) {
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, APIResponse{Success: false, Error: message})
}

// ============ API HANDLERS ============

// APIBoardsRouter роутер для /api/v1/boards/{id}
func APIBoardsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/boards/")

	if strings.Contains(path, "/threads") {
		APIGetThreads(w, r)
		return
	}

	APIGetBoard(w, r)
}

// APIGetBoards GET /api/v1/boards - получить все доски, POST - создать
func APIGetBoards(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	if r.Method == "POST" {
		APICreateBoard(w, r)
		return
	}

	boards, err := database.GetAllBoards()
	if err != nil {
		log.Printf("API: ошибка получения досок: %v", err)
		sendError(w, http.StatusInternalServerError, "Ошибка получения досок")
		return
	}

	var response []BoardResponse
	for _, b := range boards {
		response = append(response, BoardResponse{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			ThreadCount: b.ThreadCount,
			CreatedAt:   b.CreatedAt.Format(time.RFC3339),
		})
	}

	sendSuccess(w, response)
}

// APIGetBoard GET /api/v1/boards/{id} - получить доску
func APIGetBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	boardID := strings.TrimPrefix(r.URL.Path, "/api/v1/boards/")
	if boardID == "" {
		sendError(w, http.StatusBadRequest, "ID доски не указан")
		return
	}

	board, err := database.GetBoard(boardID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка получения доски")
		return
	}
	if board == nil {
		sendError(w, http.StatusNotFound, "Доска не найдена")
		return
	}

	sendSuccess(w, BoardResponse{
		ID:          board.ID,
		Name:        board.Name,
		Description: board.Description,
		CreatedAt:   board.CreatedAt.Format(time.RFC3339),
	})
}

// APICreateBoard POST /api/v1/boards - создать доску
func APICreateBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	req.ID = strings.ToLower(strings.TrimSpace(req.ID))
	req.Name = strings.TrimSpace(req.Name)

	if req.ID == "" || req.Name == "" {
		sendError(w, http.StatusBadRequest, "ID и название обязательны")
		return
	}

	// Проверка символов ID
	for _, c := range req.ID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			sendError(w, http.StatusBadRequest, "ID может содержать только латинские буквы и цифры")
			return
		}
	}

	// Проверяем существование
	existing, _ := database.GetBoard(req.ID)
	if existing != nil {
		sendError(w, http.StatusConflict, "Доска с таким ID уже существует")
		return
	}

	if err := database.CreateBoard(req.ID, req.Name, req.Description); err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка создания доски")
		return
	}

	sendSuccess(w, map[string]string{"id": req.ID, "message": "Доска создана"})
}

// APIGetThreads GET /api/v1/boards/{id}/threads - получить треды доски
func APIGetThreads(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	// Парсим путь /api/v1/boards/{boardID}/threads
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/boards/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "threads" {
		sendError(w, http.StatusBadRequest, "Неверный путь")
		return
	}
	boardID := parts[0]

	// Проверяем доску
	board, _ := database.GetBoard(boardID)
	if board == nil {
		sendError(w, http.StatusNotFound, "Доска не найдена")
		return
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "bump"
	}

	threads, err := database.GetThreadsByBoard(boardID, sortBy)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка получения тредов")
		return
	}

	var response []ThreadResponse
	for _, t := range threads {
		tr := ThreadResponse{
			ID:        t.ID,
			BoardID:   t.BoardID,
			Subject:   t.Subject,
			PostCount: t.PostCount,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
			BumpedAt:  t.BumpedAt.Format(time.RFC3339),
		}

		if t.FirstPost != nil {
			parentID := 0
			if t.FirstPost.ParentID.Valid {
				parentID = int(t.FirstPost.ParentID.Int64)
			}
			mediaPath := ""
			if t.FirstPost.MediaPath.Valid {
				mediaPath = t.FirstPost.MediaPath.String
			}
			mediaType := ""
			if t.FirstPost.MediaType.Valid {
				mediaType = t.FirstPost.MediaType.String
			}

			tr.FirstPost = &PostResponse{
				ID:        t.FirstPost.ID,
				ThreadID:  t.FirstPost.ThreadID,
				ParentID:  parentID,
				Author:    t.FirstPost.Author,
				Content:   t.FirstPost.Content,
				MediaPath: mediaPath,
				MediaType: mediaType,
				CreatedAt: t.FirstPost.CreatedAt.Format(time.RFC3339),
				Depth:     t.FirstPost.Depth,
			}
		}

		response = append(response, tr)
	}

	sendSuccess(w, response)
}

// APIGetThread GET /api/v1/threads/{id} - получить тред с постами
func APIGetThread(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	threadIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/threads/")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный ID треда")
		return
	}

	thread, err := database.GetThread(threadID)
	if err != nil || thread == nil {
		sendError(w, http.StatusNotFound, "Тред не найден")
		return
	}

	posts, err := database.GetPostsByThread(threadID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка получения постов")
		return
	}

	var postsResponse []PostResponse
	for _, p := range posts {
		parentID := 0
		if p.ParentID.Valid {
			parentID = int(p.ParentID.Int64)
		}
		mediaPath := ""
		if p.MediaPath.Valid {
			mediaPath = p.MediaPath.String
		}
		mediaType := ""
		if p.MediaType.Valid {
			mediaType = p.MediaType.String
		}

		postsResponse = append(postsResponse, PostResponse{
			ID:        p.ID,
			ThreadID:  p.ThreadID,
			ParentID:  parentID,
			Author:    p.Author,
			Content:   p.Content,
			MediaPath: mediaPath,
			MediaType: mediaType,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
			Depth:     p.Depth,
		})
	}

	response := ThreadResponse{
		ID:        thread.ID,
		BoardID:   thread.BoardID,
		Subject:   thread.Subject,
		PostCount: len(posts),
		CreatedAt: thread.CreatedAt.Format(time.RFC3339),
		BumpedAt:  thread.BumpedAt.Format(time.RFC3339),
		Posts:     postsResponse,
	}

	sendSuccess(w, response)
}

// APICreateThread POST /api/v1/threads - создать тред
func APICreateThread(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	var req struct {
		BoardID   string `json:"board_id"`
		Subject   string `json:"subject"`
		Author    string `json:"author"`
		Content   string `json:"content"`
		MediaPath string `json:"media_path"`
		MediaType string `json:"media_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	req.Subject = strings.TrimSpace(req.Subject)
	req.Content = strings.TrimSpace(req.Content)
	req.Author = strings.TrimSpace(req.Author)

	if req.BoardID == "" || req.Subject == "" || req.Content == "" {
		sendError(w, http.StatusBadRequest, "board_id, subject и content обязательны")
		return
	}

	if req.Author == "" {
		req.Author = "Аноним"
	}

	// Проверяем доску
	board, _ := database.GetBoard(req.BoardID)
	if board == nil {
		sendError(w, http.StatusNotFound, "Доска не найдена")
		return
	}

	// Создаём тред
	threadID, err := database.CreateThread(req.BoardID, req.Subject)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка создания треда")
		return
	}

	// Создаём первый пост
	postID, err := database.CreatePost(int(threadID), nil, req.Author, req.Content, req.MediaPath, req.MediaType)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка создания поста")
		return
	}

	// WebSocket уведомление
	WsHub.BroadcastToBoard(req.BoardID, WSMessage{
		Type:     "new_thread",
		ThreadID: int(threadID),
		BoardID:  req.BoardID,
	})

	sendSuccess(w, map[string]interface{}{
		"thread_id": threadID,
		"post_id":   postID,
		"message":   "Тред создан",
	})
}

// APICreatePost POST /api/v1/posts - создать пост
func APICreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	var req struct {
		ThreadID  int    `json:"thread_id"`
		ParentID  int    `json:"parent_id"`
		Author    string `json:"author"`
		Content   string `json:"content"`
		MediaPath string `json:"media_path"`
		MediaType string `json:"media_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	req.Author = strings.TrimSpace(req.Author)

	if req.ThreadID == 0 || req.Content == "" {
		sendError(w, http.StatusBadRequest, "thread_id и content обязательны")
		return
	}

	if req.Author == "" {
		req.Author = "Аноним"
	}

	// Проверяем тред
	thread, _ := database.GetThread(req.ThreadID)
	if thread == nil {
		sendError(w, http.StatusNotFound, "Тред не найден")
		return
	}

	// Создаём пост
	var parentID *int
	if req.ParentID > 0 {
		parentID = &req.ParentID
	}

	postID, err := database.CreatePost(req.ThreadID, parentID, req.Author, req.Content, req.MediaPath, req.MediaType)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка создания поста")
		return
	}

	// Бампаем тред
	database.BumpThread(req.ThreadID)

	// WebSocket уведомление
	WsHub.BroadcastToThread(req.ThreadID, WSMessage{
		Type:     "new_post",
		ThreadID: req.ThreadID,
		Data: map[string]interface{}{
			"id":         postID,
			"author":     req.Author,
			"content":    req.Content,
			"media_path": req.MediaPath,
			"media_type": req.MediaType,
			"parent_id":  req.ParentID,
			"created_at": time.Now().Format("02.01.2006 15:04:05"),
		},
	})

	WsHub.BroadcastToBoard(thread.BoardID, WSMessage{
		Type:     "thread_updated",
		ThreadID: req.ThreadID,
		BoardID:  thread.BoardID,
	})

	sendSuccess(w, map[string]interface{}{
		"post_id": postID,
		"message": "Пост создан",
	})
}

// APIUploadMedia POST /api/v1/upload - загрузить медиафайл
func (h *Handler) APIUploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		sendJSON(w, http.StatusOK, nil)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		sendError(w, http.StatusBadRequest, "Ошибка парсинга формы")
		return
	}

	fileInfo, err := saveFile(r, "media", time.Now().UnixNano())
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if fileInfo == nil {
		sendError(w, http.StatusBadRequest, "Файл не загружен")
		return
	}

	sendSuccess(w, map[string]string{
		"path": fileInfo.Path,
		"type": fileInfo.Type,
	})
}
