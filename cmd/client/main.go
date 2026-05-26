package main

import (
	"bufio" // буфферизированный ввод-вывод
	"fmt"	// форматированный вывод
	"log"	// логирование
	"net"	// сетевые операции
	"os"	// работа с оc
	"strings" // работа со строками


	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
)

const (
		serverHost = "127.0.0.1"
		serverPort = "8080"
)

func main(){
	fmt.Printf("----- Aurora -----\n")
	address := fmt.Sprintf("%s:%s", serverHost, serverPort)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close() // закрываем соединение при завершении рабоы

	log.Printf("подключено к серверу по адресу %s", address)

	scanner := bufio.NewScanner(os.Stdin) 

	var username string
	for {
		u, err := authenticate(conn, scanner)
		if err != nil {
			fmt.Printf("Ошибка входа: %v. Попробуйте снова.\n", err)
			continue
		}
		username = u
		break
	}

	fmt.Printf("Добро пожаловать, %s!\n", username)
	fmt.Println("Команды:")
	fmt.Println("  @имя сообщение — отправить сообщение пользователю")
	fmt.Printf("   /exit - выход\n")

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
		
		var packet protocol.Packet

		if strings.HasPrefix(text, "@") {
			parts := strings.SplitN(text, " ", 2)
			if len(parts) < 2 {
				fmt.Println("Формат: @имя сообщение")
				continue
			}
			to := strings.TrimPrefix(parts[0], "@")
			
			packet = protocol.Packet{
			Type:    protocol.PacketMessage,
			From:    username,
			To:      to,
			Payload: parts[1],
		}
		} else {
		fmt.Println("Используйте @имя для отправки сообщения")
		continue
	}
		

		// Отправляем пакет на сервер
		if err := protocol.Send(conn, packet); err != nil {
			log.Printf("не удалось отправить сообщение: %v", err)
			break
		}
		
		// Ждем ответ
		response, err := protocol.Receive(conn)
		if err != nil {
			log.Printf("не удалось получить ответ от сервера: %v", err)
			break
		}
		fmt.Printf("[echo]: %s: %s\n", response.From, response.Payload)
	}

}

// Функция для аутентификации пользователя
func authenticate(conn net.Conn, scanner *bufio.Scanner) (string, error) {
	var choice, username, password string

	fmt.Println("1. Вход ")
	fmt.Println("2. Регистрация ")
	fmt.Println("Выберите действие: ")

	scanner.Scan()
	choice = strings.TrimSpace(scanner.Text())

	for {
		fmt.Print("Имя пользователя: ")
		scanner.Scan()
		username = strings.TrimSpace(scanner.Text())
		if username != "" {
			break
		}
		fmt.Println("Ошибка: имя пользователя не может быть пустым")
	}

	for {
		fmt.Print("Пароль: ")
		scanner.Scan()
		password = strings.TrimSpace(scanner.Text())
		if password != "" {
			break
		}
		fmt.Println("Ошибка: пароль не может быть пустым")
	}

	var packetType string
	switch choice {
	case "1":
		packetType = protocol.PacketLogin
	case "2":
		packetType = protocol.PacketRegister
	default:
		return "", fmt.Errorf("недопустимое значение: %s", choice)
	}

	// Отправляем пакет аутентификации на сервер
	err := protocol.Send(conn, protocol.Packet{
		Type:    packetType,
		From:    username,
		Payload: password,
	})
	if err != nil {
		return "", fmt.Errorf("не удалось отправить пакет аутентификации: %w", err)
	}

	// Ждём ответа от сервера
	response, err := protocol.Receive(conn)
	if err != nil {
		return "", fmt.Errorf("не удалось получить ответ от сервера: %w", err)
	}

	if response.Type == protocol.PacketError {
		return "", fmt.Errorf(response.Payload)
	}

	return username, nil
}