package chat

import (
    "fmt"
    "net"
    "sync"

    "github.com/fulikozzz/AuroraMsgr/internal/protocol"
)

// Подключенный пользователь
type Client struct {
    Username string
    Conn net.Conn
}

// Набор клиентов
type Hub struct {
    mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
    return &Hub{
        clients: make(map[string]*Client),
    }
}

