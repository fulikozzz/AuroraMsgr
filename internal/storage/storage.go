package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite", path) // открываем БД
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть БД: %w", err)
	}

	s := &Storage{db: db} // создаём структуру для работы с БД
	if err := s.migrate(); err != nil { // создаём таблицы если их нет
		return nil, fmt.Errorf("не удалось создать таблицы: %w", err)
	}

	return s, nil
}

// функция создаёт таблицы
func (s *Storage) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT    NOT NULL UNIQUE,
		password TEXT    NOT NULL,
		created_at INTEGER  NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		from_user  TEXT NOT NULL,
		to_user    TEXT NOT NULL,
		payload    TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);`

	// выполняем SQL запрос для создания таблиц
	_, err := s.db.Exec(query) 
	return err
}

// Закрывает соединение с БД
func (s *Storage) Close() {
	s.db.Close()
}

// Сохраняет нового пользователя
func (s *Storage) CreateUser(username, password string) error {
	_, err := s.db.Exec(
		"INSERT INTO users (username, password, created_at) VALUES (?, ?, ?)",
		username, password, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("пользователь уже существует")
	}
	return nil
}

// Возвращает пользователя по имени
func (s *Storage) GetUser(username string) (string, string, error) {
	var user, password string
	err := s.db.QueryRow(
		"SELECT username, password FROM users WHERE username = ?",
		username,
	).Scan(&user, &password)
	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("пользователь не найден")
	}
	if err != nil {
		return "", "", fmt.Errorf("ошибка БД: %w", err)
	}
	return user, password, nil
}

// Сохраняет сообщение в БД
func (s *Storage) SaveMessage(from, to, payload string) error {
	_, err := s.db.Exec(
		"INSERT INTO messages (from_user, to_user, payload, created_at) VALUES (?, ?, ?, ?)",
		from, to, payload, time.Now().Unix(),
	)
	return err
}

// Возвращает историю сообщений между двумя пользователями
func (s *Storage) GetHistory(user1, user2 string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT from_user, to_user, payload, created_at 
		FROM messages 
		WHERE (from_user = ? AND to_user = ?) 
		   OR (from_user = ? AND to_user = ?)
		ORDER BY created_at ASC`,
		user1, user2, user2, user1,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var from, to, payload, createdAt string
		if err := rows.Scan(&from, &to, &payload, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, fmt.Sprintf("[%s] %s -> %s: %s", createdAt, from, to, payload))
	}
	return result, nil
}