package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"webForum/database"
	"webForum/handlers"
)

func main() {
	// Загрузка .env файла (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения системы")
	}
	// Конфигурация БД(можно вынести в переменные окружения)
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "webforum"),
	}

	// Подключение к MySQL
	if err := database.Connect(dbConfig); err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	defer database.Close()

	// Создание таблиц
	if err := database.InitSchema(); err != nil {
		log.Fatal("Ошибка инициализации схемы: ", err)
	}

	// Настройка маршрутизатора
	mux := http.NewServeMux()

	// Статические файлы (CSS, JS, изображения)
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Загруженные файлы пользователей
	imgFs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", imgFs))

	// Инициализация обработчиков
	h := handlers.NewHandler()

	// === СТРАНИЦЫ ===
	// Главная страница - список всех досок
	mux.HandleFunc("/", h.IndexHandler)

	// Страница доски - список тредов
	// GET /board/{id}?sort=bump|new|old|replies
	mux.HandleFunc("/board/", h.BoardHandler)

	// Страница треда - список комментариев
	// GET /thread/{id}
	mux.HandleFunc("/thread/", h.ThreadHandler)

	// === API (POST запросы) ===
	// Создание новой доски
	// POST /api/board  {id, name, description}
	mux.HandleFunc("/api/board", h.CreateBoardHandler)

	// Создание нового треда
	// POST /api/thread  {board_id, subject, author, content, image}
	mux.HandleFunc("/api/thread", h.CreateThreadHandler)

	// Создание нового поста/комментария
	// POST /api/post  {thread_id, parent_id, author, content, image}
	mux.HandleFunc("/api/post", h.CreatePostHandler)

	// === WebSocket ===
	mux.HandleFunc("/ws/thread", handlers.WebSocketThreadHandler)
	mux.HandleFunc("/ws/board", handlers.WebSocketBoardHandler)

	// Запуск сервера
	port := getEnv("PORT", ":8080")
	if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("Сервер запущен на http://localhost%s", port)
	log.Println("Маршруты:")
	log.Println("  GET  /              - Главная страница")
	log.Println("  GET  /board/{id}    - Страница доски")
	log.Println("  GET  /thread/{id}   - Страница треда")
	log.Println("  POST /api/board     - Создать доску")
	log.Println("  POST /api/thread    - Создать тред")
	log.Println("  POST /api/post      - Создать пост")

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}

// getEnv возвращает переменную окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
