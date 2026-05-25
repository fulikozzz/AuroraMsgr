package auth

import (
    "errors"    // работа с ошибками

    "github.com/fulikozzz/AuroraMsgr/internal/storage"
)


type Manager struct {
    storage *storage.Storage
}

// Создание нового менеджера аутентификации
func NewManager(s *storage.Storage) *Manager {
    return &Manager{
        storage: s,
    }
}

// Регистрация нового пользователя
func (m *Manager) Register(username, password string) error {
    if username == "" || password == "" {
        return errors.New("имя пользователя и пароль не могут быть пустыми")
    }

    if err := m.storage.CreateUser(username, password); err != nil {
        return err
    }

    return nil
}

func (m *Manager) Login (username, password string) error {
    _, storedPassword, err := m.storage.GetUser(username)

    if err != nil {
        return errors.New("пользователь не найден")
    }

    if storedPassword != password {
        return errors.New("неверный пароль")
    }
    
    return nil
}
