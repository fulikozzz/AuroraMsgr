package auth

import (
    "errors"    // работа с ошибками
    "sync"      // синхронизация доступа
)

type User struct {
    Username string
    Password string
}

type Manager struct {
    mu sync.RWMutex
    users map[string]User
}

// Создание нового менеджера аутентификации
func NewManager() *Manager {
    return &Manager{
        users: make(map[string]User),
    }
}

// Регистрация нового пользователя
func (m *Manager) Register(username, password string) error {
    m.mu.Lock() // блокируем для записи 
    defer m.mu.Unlock() // разблокируем после записи

    if username == "" || password == "" {
        return errors.New("имя пользователя и пароль не могут быть пустыми")
    }

    if _, exists := m.users[username]; exists {
        return errors.New("пользователь уже существует")
    }

    m.users[username] = User{
        Username: username,
        Password: password,
    }

    return nil
}

func (m *Manager) Login (username, password string) error {
    m.mu.RLock() // блокируем для чтения
    defer m.mu.RUnlock() // разблокируем после чтения
    
    user, exists := m.users[username]
   
    if !exists {
        return errors.New("пользователь не найден")
    }

    if user.Password != password {
        return errors.New("неверный пароль")
    }
    
    return nil
}

func (m *Manager) Exists(username string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()

    _, exists := m.users[username]
    return exists
}
