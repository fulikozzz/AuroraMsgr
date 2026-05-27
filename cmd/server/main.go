package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
	"github.com/fulikozzz/AuroraMsgr/internal/auth"
	"github.com/fulikozzz/AuroraMsgr/internal/storage"
	"github.com/fulikozzz/AuroraMsgr/internal/logger"
	"github.com/fulikozzz/AuroraMsgr/internal/chat"
)

// Конфигурация сервера
const  (
	host = "0.0.0.0"
	port = "8080"
)

type Server struct {
	authManager *auth.Manager
	storage		*storage.Storage
	hub			*chat.Hub
	log			*logger.Logger
}

func NewServer(storage *storage.Storage, log *logger.Logger) *Server {
	return &Server{
		authManager:	auth.NewManager(storage),
		storage:		storage,
		hub:			chat.NewHub(),
		log:			log,
	}
}

func main(){
	serverLogger, err := logger.NewLogger("server")
	if err != nil {
		serverLogger.Error("не удалось создать логгер: %v", err)
	}
	defer serverLogger.Close()

	database, err := storage.NewStorage("aurora.db")
	if err != nil {
		serverLogger.Error("не удалось открыть БД: %v", err)
		return
	}
	defer database.Close()

	server := NewServer( database, serverLogger)
	address := fmt.Sprintf("%s:%s", host, port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		serverLogger.Error("не удалось запустить сервер: %v", err)
	}
	defer listener.Close()

	serverLogger.Info("сервер AuroraMsgr по адресу %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			serverLogger.Error("не удалось принять соединение: %v", err)
			continue
		}

		serverLogger.Info("новое соединение от %s", conn.RemoteAddr())
		go server.handleConnection(conn)
	}
}

// Обработка соединения
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Первый пакет должен быть запросом на аутентификацию
	username, err := s.authenticate(conn)
	if err != nil {
		s.log.Error("аутентификация не удалась для %s: %v", conn.RemoteAddr(), err)
		return
	}

	s.log.Auth("пользователь аутентифицирован", username)
	
	// Регаем клиента в Hub
	s.hub.Register(username, conn)
	defer s.hub.Unregister(username)

	s.log.Info("пользователь %s онлайн", username)

	protocol.Send(conn, protocol.Packet{
		Type: protocol.PacketSuccess,
		Payload: fmt.Sprintf("Добро пожаловать, %s!", username),
	})

	// Основной цикл обработки сообщений от клиента
	for {
		packet, err := protocol.Receive(conn)
		if err != nil {
			s.log.Auth(username, "отключился")
			return
		}

		// Обработка команды /online
		if packet.Payload == "/online" {
			users := s.hub.GetOnlineUsers()
			protocol.Send(conn, protocol.Packet{
				Type: protocol.PacketSystem,
				Payload: fmt.Sprintf("Пользователи онлайн: %s", strings.Join(users, ", ")),
			})
			
			s.log.Info("пользователь %s запросил список онлайн пользователей", username)
			continue
		}

		s.log.Msg(packet.From, packet.To, packet.Payload)
		s.storage.SaveMessage(packet.From, packet.To, packet.Payload)

		if err := s.hub.Send(packet.From, packet.To, packet.Payload); err != nil {
			s.log.Error("не удалось отправить пакет: %v", err)
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
			s.log.Auth(packet.From, "зарегистрировался")
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
			s.log.Auth(packet.From, "вошел в сеть")
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