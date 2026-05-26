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

// Регистрирует клиента в Hub
func (h *Hub) Register(username string, conn net.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()

    h.clients[username] = &Client{
        Username: username,
        Conn: conn,
    }
}

// Удаляет клиента из Hub
func (h *Hub) Unregister(username string){
    h.mu.Lock()
    defer h.mu.Unlock()

    delete(h.clients, username)
}

// Отправка сообщения пользовател.ю
func (h *Hub) Send(from, to, payload string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.clients[to]
	if !exists {
		return fmt.Errorf("пользователь '%s' не в сети", to)
	}

	return protocol.Send(client.Conn, protocol.Packet{
		Type:    protocol.PacketMessage,
		From:    from,
		To:      to,
		Payload: payload,
	})
}