package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"webForum/database"
)

// Допустимые типы файлов
var allowedExtensions = map[string]string{
	// Изображения
	".jpg":  "image",
	".jpeg": "image",
	".png":  "image",
	".gif":  "image",
	".webp": "image",
	".svg":  "image",
	// Видео
	".mp4":  "video",
	".webm": "video",
	".avi":  "video",
	".mov":  "video",
	".mkv":  "video",
	// Аудио
	".mp3":  "audio",
	".wav":  "audio",
	".ogg":  "audio",
	".flac": "audio",
	".m4a":  "audio",
}

// FileInfo информация о загруженном файле
type FileInfo struct {
	Path string
	Type string
}

// saveFile сохраняет загруженный файл
func saveFile(r *http.Request, fieldName string, postID int64) (*FileInfo, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	fileType, ok := allowedExtensions[ext]
	if !ok {
		return nil, fmt.Errorf("недопустимый тип файла: %s", ext)
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%d_%d%s", postID, time.Now().UnixNano(), ext)
	filePath := filepath.Join("uploads", filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	return &FileInfo{
		Path: "/uploads/" + filename,
		Type: fileType,
	}, nil
}

type Handler struct {
	templates *template.Template
}

func NewHandler() *Handler {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("02.01.2006 15:04:05")
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"nullStr": func(ns sql.NullString) string {
			if ns.Valid {
				return ns.String
			}
			return ""
		},
		"nullInt": func(ni sql.NullInt64) int {
			if ni.Valid {
				return int(ni.Int64)
			}
			return 0
		},
	}).ParseGlob("templates/*.html"))

	return &Handler{
		templates: tmpl,
	}
}

// IndexHandler - главная страница со списком досок
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	boards, err := database.GetAllBoards()
	if err != nil {
		log.Printf("Ошибка получения досок: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":  "Веб-форум",
		"Boards": boards,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Ошибка рендеринга: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// BoardHandler - страница доски с тредами
func (h *Handler) BoardHandler(w http.ResponseWriter, r *http.Request) {
	boardID := strings.TrimPrefix(r.URL.Path, "/board/")
	if boardID == "" {
		http.NotFound(w, r)
		return
	}

	board, err := database.GetBoard(boardID)
	if err != nil {
		log.Printf("Ошибка получения доски: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if board == nil {
		http.NotFound(w, r)
		return
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "bump"
	}

	threads, err := database.GetThreadsByBoard(boardID, sortBy)
	if err != nil {
		log.Printf("Ошибка получения тредов: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   "/" + boardID + "/ - " + board.Name,
		"Board":   board,
		"BoardID": boardID,
		"SortBy":  sortBy,
		"Threads": threads,
	}

	if err := h.templates.ExecuteTemplate(w, "board.html", data); err != nil {
		log.Printf("Ошибка рендеринга: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// ThreadHandler - страница треда с комментариями
func (h *Handler) ThreadHandler(w http.ResponseWriter, r *http.Request) {
	threadIDStr := strings.TrimPrefix(r.URL.Path, "/thread/")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	thread, err := database.GetThread(threadID)
	if err != nil {
		log.Printf("Ошибка получения треда: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if thread == nil {
		http.NotFound(w, r)
		return
	}

	board, err := database.GetBoard(thread.BoardID)
	if err != nil {
		log.Printf("Ошибка получения доски: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	posts, err := database.GetPostsByThread(threadID)
	if err != nil {
		log.Printf("Ошибка получения постов: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":    thread.Subject,
		"Thread":   thread,
		"ThreadID": threadID,
		"Board":    board,
		"BoardID":  thread.BoardID,
		"Posts":    posts,
	}

	if err := h.templates.ExecuteTemplate(w, "thread.html", data); err != nil {
		log.Printf("Ошибка рендеринга: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// CreateBoardHandler - создание новой доски
func (h *Handler) CreateBoardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	id := strings.ToLower(strings.TrimSpace(r.FormValue("id")))
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))

	if id == "" || name == "" {
		http.Error(w, "ID и название обязательны", http.StatusBadRequest)
		return
	}

	// Проверяем допустимые символы
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			http.Error(w, "ID может содержать только латинские буквы и цифры", http.StatusBadRequest)
			return
		}
	}

	// Проверяем существование
	existing, _ := database.GetBoard(id)
	if existing != nil {
		http.Error(w, "Доска с таким ID уже существует", http.StatusBadRequest)
		return
	}

	if err := database.CreateBoard(id, name, description); err != nil {
		log.Printf("Ошибка создания доски: %v", err)
		http.Error(w, "Ошибка создания доски", http.StatusInternalServerError)
		return
	}

	log.Printf("✓ Создана доска: /%s/ - %s", id, name)
	http.Redirect(w, r, "/board/"+id, http.StatusSeeOther)
}

// CreateThreadHandler - создание нового треда
func (h *Handler) CreateThreadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	boardID := r.FormValue("board_id")
	subject := strings.TrimSpace(r.FormValue("subject"))
	author := strings.TrimSpace(r.FormValue("author"))
	content := strings.TrimSpace(r.FormValue("content"))

	if subject == "" || content == "" {
		http.Error(w, "Тема и сообщение обязательны", http.StatusBadRequest)
		return
	}

	if author == "" {
		author = "Аноним"
	}

	// Проверяем существование доски
	board, _ := database.GetBoard(boardID)
	if board == nil {
		http.Error(w, "Доска не найдена", http.StatusNotFound)
		return
	}

	// Создаём тред
	threadID, err := database.CreateThread(boardID, subject)
	if err != nil {
		log.Printf("Ошибка создания треда: %v", err)
		http.Error(w, "Ошибка создания треда", http.StatusInternalServerError)
		return
	}

	// Сохраняем медиафайл
	fileInfo, err := saveFile(r, "media", threadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mediaPath := ""
	mediaType := ""
	if fileInfo != nil {
		mediaPath = fileInfo.Path
		mediaType = fileInfo.Type
	}

	// Создаём первый пост (OP)
	_, err = database.CreatePost(int(threadID), nil, author, content, mediaPath, mediaType)
	if err != nil {
		log.Printf("Ошибка создания поста: %v", err)
		http.Error(w, "Ошибка создания поста", http.StatusInternalServerError)
		return
	}

	log.Printf("✓ Создан тред #%d: %s", threadID, subject)
	http.Redirect(w, r, "/thread/"+strconv.FormatInt(threadID, 10), http.StatusSeeOther)
}

// CreatePostHandler - создание нового поста
func (h *Handler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	threadIDStr := r.FormValue("thread_id")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "Неверный ID треда", http.StatusBadRequest)
		return
	}

	parentIDStr := r.FormValue("parent_id")
	var parentID *int
	if parentIDStr != "" && parentIDStr != "0" {
		pid, _ := strconv.Atoi(parentIDStr)
		if pid > 0 {
			parentID = &pid
		}
	}

	author := strings.TrimSpace(r.FormValue("author"))
	content := strings.TrimSpace(r.FormValue("content"))

	if content == "" {
		http.Error(w, "Сообщение обязательно", http.StatusBadRequest)
		return
	}

	if author == "" {
		author = "Аноним"
	}

	// Проверяем существование треда
	thread, _ := database.GetThread(threadID)
	if thread == nil {
		http.Error(w, "Тред не найден", http.StatusNotFound)
		return
	}

	// Сохраняем медиафайл
	fileInfo, err := saveFile(r, "media", int64(threadID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mediaPath := ""
	mediaType := ""
	if fileInfo != nil {
		mediaPath = fileInfo.Path
		mediaType = fileInfo.Type
	}

	// Создаём пост
	postID, err := database.CreatePost(threadID, parentID, author, content, mediaPath, mediaType)
	if err != nil {
		log.Printf("Ошибка создания поста: %v", err)
		http.Error(w, "Ошибка создания поста", http.StatusInternalServerError)
		return
	}

	// Бампаем тред
	database.BumpThread(threadID)

	// Отправляем WebSocket уведомление
	parentIDVal := 0
	if parentID != nil {
		parentIDVal = *parentID
	}
	WsHub.BroadcastToThread(threadID, WSMessage{
		Type:     "new_post",
		ThreadID: threadID,
		Data: map[string]interface{}{
			"id":         postID,
			"author":     author,
			"content":    content,
			"media_path": mediaPath,
			"media_type": mediaType,
			"parent_id":  parentIDVal,
			"created_at": time.Now().Format("02.01.2006 15:04:05"),
		},
	})

	// Также уведомляем доску о новом посте
	WsHub.BroadcastToBoard(thread.BoardID, WSMessage{
		Type:     "thread_updated",
		ThreadID: threadID,
		BoardID:  thread.BoardID,
	})

	log.Printf("✓ Создан пост #%d в треде #%d", postID, threadID)
	http.Redirect(w, r, "/thread/"+threadIDStr+"#post-"+strconv.FormatInt(postID, 10), http.StatusSeeOther)
}
