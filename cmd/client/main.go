package main

import (
	"bufio" // буфферизированный ввод-вывод
	"fmt"	// форматированный вывод
	"log"	// логирование
	"net"	// сетевые операции
	"os"	// работа с оc

	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
)

const (
		serverHost = "127.0.0.1"
		serverPort = "8080"
)

func main(){
	address := fmt.Sprintf("%s:%s", serverHost, serverPort)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close() // закрываем соединение при завершении рабоы

	log.Printf("подключено к серверу по адресу %s", address)

	scanner := bufio.NewScanner(os.Stdin) 

	fmt.Print("Введите ваше имя: ")
	scanner.Scan()
	name := scanner.Text()

	// Запускаем горутину для получения сообщений от сервера
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		text := scanner.Text()
		if text == "" {
			continue
		}
		
		if text == "/exit" {
			fmt.Println("отключение от сервера...")
			break
		}

		packet := protocol.Packet{
			Type: protocol.PacketMessage,
			From: name,
			To: "server",
			Payload: text,
		}

		// Отправляем пакет на сервер
		if err := protocol.Send(conn, packet); err != nil {
			log.Printf("не удалось отправить сообщение: %v", err)
			break
		}
		
		// Получаем ответ от сервера
		response, err := protocol.Receive(conn)
		if err != nil {
			log.Printf("не удалось получить ответ от сервера: %v", err)
			break
		}
		fmt.Printf("[echo]: %s: %s\n", response.From, response.Payload)
	}

}
