package main

import (
	"fmt"
	"log"
	"net"

	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
)

// Конфигурация сервера
const  (
	host = "0.0.0.0"
	port = "8080"
)


func main(){
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
		go handleConnection(conn)
	}
}

// Обработка соединения
func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		packet, err := protocol.Receive(conn)
		if err != nil {
			log.Printf("соединение разорвано: %s", conn.RemoteAddr())
			return
		}

		log.Printf("[%s -> %s]: %s", packet.From, packet.To, packet.Payload)

		if err := protocol.Send(conn, packet); err != nil {
			log.Printf("не удалось отправить пакет: %v", err)
			return
		}
	}
}