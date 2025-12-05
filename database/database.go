package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Config конфигурация подключения к БД
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// Connect подключается к MySQL
func Connect(cfg Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("ошибка открытия соединения: %w", err)
	}
	
	// Настройка пула соединений
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
	
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	
	log.Println("✓ Подключение к MySQL успешно")
	return nil
}

// Close закрывает соединение с БД
func Close() {
	if DB != nil {
		DB.Close()
		log.Println("Соединение с БД закрыто")
	}
}

// InitSchema создаёт таблицы если их нет
func InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS boards (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		`CREATE TABLE IF NOT EXISTS threads (
			id INT AUTO_INCREMENT PRIMARY KEY,
			board_id VARCHAR(50) NOT NULL,
			subject VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			bumped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
			INDEX idx_board_bumped (board_id, bumped_at DESC)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		
		`CREATE TABLE IF NOT EXISTS posts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			thread_id INT NOT NULL,
			parent_id INT DEFAULT NULL,
			author VARCHAR(100) DEFAULT 'Аноним',
			content TEXT NOT NULL,
			media_path VARCHAR(500),
			media_type VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES posts(id) ON DELETE SET NULL,
			INDEX idx_thread (thread_id),
			INDEX idx_parent (parent_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	
	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("ошибка создания таблицы: %w", err)
		}
	}
	
	log.Println("✓ Таблицы созданы/проверены")
	return nil
}
