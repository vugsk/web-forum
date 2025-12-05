package database

import (
	"database/sql"
	"time"
)

// Board доска форума
type Board struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	ThreadCount int // вычисляемое поле
}

// Thread тред на доске
type Thread struct {
	ID        int
	BoardID   string
	Subject   string
	CreatedAt time.Time
	BumpedAt  time.Time
	PostCount int   // вычисляемое поле
	FirstPost *Post // первый пост (OP)
}

// Post пост/комментарий в треде
type Post struct {
	ID        int
	ThreadID  int
	ParentID  sql.NullInt64
	Author    string
	Content   string
	MediaPath sql.NullString
	MediaType sql.NullString
	CreatedAt time.Time
	Depth     int // глубина вложенности для лесенки
}

// === BOARDS ===

// GetAllBoards возвращает все доски
func GetAllBoards() ([]Board, error) {
	query := `
		SELECT b.id, b.name, b.description, b.created_at,
		       COALESCE(COUNT(t.id), 0) as thread_count
		FROM boards b
		LEFT JOIN threads t ON b.id = t.board_id
		GROUP BY b.id
		ORDER BY b.id`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var boards []Board
	for rows.Next() {
		var b Board
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt, &b.ThreadCount); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, nil
}

// GetBoard возвращает доску по ID
func GetBoard(id string) (*Board, error) {
	query := `SELECT id, name, description, created_at FROM boards WHERE id = ?`
	
	var b Board
	err := DB.QueryRow(query, id).Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBoard создаёт новую доску
func CreateBoard(id, name, description string) error {
	query := `INSERT INTO boards (id, name, description) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, id, name, description)
	return err
}

// === THREADS ===

// GetThreadsByBoard возвращает треды доски с сортировкой
func GetThreadsByBoard(boardID, sortBy string) ([]Thread, error) {
	var orderBy string
	switch sortBy {
	case "new":
		orderBy = "t.created_at DESC"
	case "old":
		orderBy = "t.created_at ASC"
	case "replies":
		orderBy = "post_count DESC"
	default: // bump
		orderBy = "t.bumped_at DESC"
	}
	
	query := `
		SELECT t.id, t.board_id, t.subject, t.created_at, t.bumped_at,
		       COUNT(p.id) as post_count
		FROM threads t
		LEFT JOIN posts p ON t.id = p.thread_id
		WHERE t.board_id = ?
		GROUP BY t.id
		ORDER BY ` + orderBy
	
	rows, err := DB.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var threads []Thread
	for rows.Next() {
		var t Thread
		if err := rows.Scan(&t.ID, &t.BoardID, &t.Subject, &t.CreatedAt, &t.BumpedAt, &t.PostCount); err != nil {
			return nil, err
		}
		
		// Получаем первый пост (OP)
		t.FirstPost, _ = GetFirstPost(t.ID)
		threads = append(threads, t)
	}
	return threads, nil
}

// GetThread возвращает тред по ID
func GetThread(id int) (*Thread, error) {
	query := `SELECT id, board_id, subject, created_at, bumped_at FROM threads WHERE id = ?`
	
	var t Thread
	err := DB.QueryRow(query, id).Scan(&t.ID, &t.BoardID, &t.Subject, &t.CreatedAt, &t.BumpedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateThread создаёт новый тред и возвращает его ID
func CreateThread(boardID, subject string) (int64, error) {
	query := `INSERT INTO threads (board_id, subject) VALUES (?, ?)`
	result, err := DB.Exec(query, boardID, subject)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// BumpThread обновляет время последнего бампа
func BumpThread(threadID int) error {
	query := `UPDATE threads SET bumped_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := DB.Exec(query, threadID)
	return err
}

// === POSTS ===

// GetPostsByThread возвращает все посты треда
func GetPostsByThread(threadID int) ([]Post, error) {
	query := `
		SELECT id, thread_id, parent_id, author, content, media_path, media_type, created_at
		FROM posts
		WHERE thread_id = ?
		ORDER BY created_at ASC`
	
	rows, err := DB.Query(query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.ThreadID, &p.ParentID, &p.Author, &p.Content,
			&p.MediaPath, &p.MediaType, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	
	// Вычисляем глубину для лесенки
	return buildPostTree(posts), nil
}

// GetFirstPost возвращает первый пост треда (OP)
func GetFirstPost(threadID int) (*Post, error) {
	query := `
		SELECT id, thread_id, parent_id, author, content, media_path, media_type, created_at
		FROM posts
		WHERE thread_id = ?
		ORDER BY created_at ASC
		LIMIT 1`
	
	var p Post
	err := DB.QueryRow(query, threadID).Scan(&p.ID, &p.ThreadID, &p.ParentID, &p.Author,
		&p.Content, &p.MediaPath, &p.MediaType, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreatePost создаёт новый пост и возвращает его ID
func CreatePost(threadID int, parentID *int, author, content, mediaPath, mediaType string) (int64, error) {
	var query string
	var result sql.Result
	var err error
	
	if parentID != nil && *parentID > 0 {
		query = `INSERT INTO posts (thread_id, parent_id, author, content, media_path, media_type) VALUES (?, ?, ?, ?, ?, ?)`
		result, err = DB.Exec(query, threadID, *parentID, author, content, nullString(mediaPath), nullString(mediaType))
	} else {
		query = `INSERT INTO posts (thread_id, author, content, media_path, media_type) VALUES (?, ?, ?, ?, ?)`
		result, err = DB.Exec(query, threadID, author, content, nullString(mediaPath), nullString(mediaType))
	}
	
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// === HELPERS ===

// nullString возвращает nil для пустых строк
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// buildPostTree вычисляет глубину вложенности постов
func buildPostTree(posts []Post) []Post {
	if len(posts) == 0 {
		return posts
	}
	
	// Создаём карту для быстрого поиска
	postMap := make(map[int]*Post)
	for i := range posts {
		postMap[posts[i].ID] = &posts[i]
	}
	
	// Вычисляем глубину
	var result []Post
	var addWithDepth func(post *Post, depth int)
	
	processed := make(map[int]bool)
	
	addWithDepth = func(post *Post, depth int) {
		if processed[post.ID] {
			return
		}
		processed[post.ID] = true
		post.Depth = depth
		result = append(result, *post)
		
		// Добавляем дочерние посты
		for i := range posts {
			if posts[i].ParentID.Valid && int(posts[i].ParentID.Int64) == post.ID {
				addWithDepth(&posts[i], depth+1)
			}
		}
	}
	
	// Начинаем с корневых постов (без parent)
	for i := range posts {
		if !posts[i].ParentID.Valid {
			addWithDepth(&posts[i], 0)
		}
	}
	
	// Добавляем посты с несуществующим parent
	for i := range posts {
		if !processed[posts[i].ID] {
			posts[i].Depth = 0
			result = append(result, posts[i])
		}
	}
	
	return result
}
