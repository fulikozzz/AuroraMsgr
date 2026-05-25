package main

import (
	"fmt"
	"log"
	"net"

	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
	"github.com/fulikozzz/AuroraMsgr/internal/auth"
)

// Конфигурация сервера
const  (
	host = "0.0.0.0"
	port = "8080"
)

type Server struct {
	authManager *auth.Manager
}

func NewServer() *Server {
	return &Server{
		authManager: auth.NewManager(),
	}
}

func main(){
	server := NewServer()
	address := fmt.Sprintf("%s:%s", host, port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
	defer listener.Close()

	log.Printf("сервер AuroraMsgr по адресу %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("не удалось принять соединение: %v", err)
			continue
		}

		log.Printf("новое соединение от %s", conn.RemoteAddr())
		go server.handleConnection(conn)
	}
}

// Обработка соединения
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Первый пакет должен быть запросом на аутентификацию
	username, err := s.authenticate(conn)
	if err != nil {
		log.Printf("аутентификация не удалась для %s: %v", conn.RemoteAddr(), err)
		return
	}

	log.Printf("пользователь %s аутентифицирован", username)
	
	protocol.Send(conn, protocol.Packet{
		Type: protocol.PacketSuccess,
		Payload: fmt.Sprintf("Добро пожаловать, %s!", username),
	})

	for {
		packet, err := protocol.Receive(conn)
		if err != nil {
			log.Printf("пользователь %s отключился: %v", username, err)
			return
		}

		log.Printf("[%s -> %s]: %s", packet.From, packet.To, packet.Payload)

		if err := protocol.Send(conn, packet); err != nil {
			log.Printf("не удалось отправить пакет: %v", err)
			return
		}
	}
}

func (s *Server) authenticate(conn net.Conn) (string, error) {
	for {
		packet, err := protocol.Receive(conn)
		if err != nil {
			return "", fmt.Errorf("не удалось получить пакет аутентификации: %w", err)
		}

		switch packet.Type {
		case protocol.PacketRegister:
			if err := s.authManager.Register(packet.From, packet.Payload); err != nil {
				protocol.Send(conn, protocol.Packet{
					Type:    protocol.PacketError,
					Payload: err.Error(),
				})
				// Повторяем попытку
				continue
			}
			return packet.From, nil

		case protocol.PacketLogin:
			if err := s.authManager.Login(packet.From, packet.Payload); err != nil {
				protocol.Send(conn, protocol.Packet{
					Type:    protocol.PacketError,
					Payload: err.Error(),
				})
				// Повторяем попытку
				continue
			}
			return packet.From, nil

		default:
			protocol.Send(conn, protocol.Packet{
				Type:    protocol.PacketError,
				Payload: "первый пакет должен быть Login или Register",
			})
			// Повторяем попытку
			continue
		}
	}
}