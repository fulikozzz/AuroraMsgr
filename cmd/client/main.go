package main

import (
	"bufio" // буфферизированный ввод-вывод
	"fmt"   // форматированный вывод
	"log"   // логирование
	"net"   // сетевые операции
	"os"    // работа с оc
	"strings" // работа со строками
	"time"


	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"

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

	runInterface(conn, username)
}

func runInterface(conn net.Conn, username  string){
	app := tview.NewApplication()

	// Область сообщений
	messages := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	messages.SetBorder(true).SetTitle(" AuroraMsgr ")

	// Поле ввода
	input := tview.NewInputField().
		SetLabel("> ").
		SetFieldBackgroundColor(0)
	input.SetBorder(true)

	// Подсказки
	hints := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	hints.SetBorder(true).SetTitle(" Подсказки ")
	hintsText := `[yellow]Команды:[-]
[@имя сообщение] — отправить личное сообщение
/online — показать онлайн пользователей
/exit — выйти из приложения
[green]Навигация:[-]
[F1] — Онлайн
[Ctrl+C] — Выход

[white]Пример:[-] @ivan Привет!`
	hints.SetText(hintsText)
	hints.SetTextAlign(tview.AlignLeft)

	// Список онлайн пользователей (можно выбрать, чтобы подставить в поле ввода)
	onlineList := tview.NewList()
	onlineList.SetBorder(true)
	onlineList.SetTitle(" Онлайн ")

	onlineList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
    input.SetText("@" + strings.TrimSpace(mainText) + " ")
    app.SetFocus(input)
	})

	// Правый столбец: подсказки сверху, список онлайн снизу
	rightFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(hints, 0, 1, false).
		AddItem(onlineList, 0, 2, false)

	// Горизонтальный контейнер: слева — сообщения, справа — подсказки/онлайн
	contentFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(messages, 0, 1, false).  // растягиваемая область сообщений
		AddItem(rightFlex, 35, 0, false)    // фиксированная ширина правой колонки

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, false).
		AddItem(input, 3, 0, true)

	// Запрашиваем список онлайн сразу при входе
	protocol.Send(conn, protocol.Packet{
		Type:    protocol.PacketMessage,
		From:    username,
		To:      "server",
		Payload: "/online",
	})
	
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			protocol.Send(conn, protocol.Packet{
				Type:    protocol.PacketMessage,
				From:    username,
				To:      "server",
				Payload: "/online",
			})
		}
	}()

	// Горутина для получения сообщений от сервера
	go func() {
		for {
			packet, err := protocol.Receive(conn)
			if err != nil {
				fmt.Println("\nсоединение с сервером потеряно")
				os.Exit(0)
			}
			switch packet.Type {
			case protocol.PacketMessage:
				// Добавляем отметку времени
				ts := time.Now().Format("15:04")
				fmt.Fprintf(messages, "[gray]%s[-] [green]%s[-]: %s\n", ts, packet.From, packet.Payload)
				messages.ScrollToEnd()
			case protocol.PacketSystem:
				// Обновление списка онлайн при получении соответствующего системного сообщения
				const prefix = "Пользователи онлайн:"
				if strings.HasPrefix(packet.Payload, prefix) {
					list := strings.TrimSpace(strings.TrimPrefix(packet.Payload, prefix))
					onlineList.Clear()
					if list != "" {
						users := strings.Split(list, ",")
						for _, u := range users {
							u = strings.TrimSpace(u)
							if u == "" {
								continue
							}
							onlineList.AddItem(u, "", 0, nil)
						}
					}
					// Также показываем краткое системное сообщение в чат
					fmt.Fprintf(messages, "[yellow]%s[-]\n", packet.Payload)
					messages.ScrollToEnd()
					app.Draw()
					continue
				}
				fmt.Fprintf(messages, "[yellow]%s[-]\n", packet.Payload)
				messages.ScrollToEnd()
			case protocol.PacketError:
				fmt.Fprintf(messages, "[red]Ошибка: %s[-]\n", packet.Payload)
				messages.ScrollToEnd()
			}
		}
	}()

	// Цикл отправки сообщений
	input.SetDoneFunc(func(key tcell.Key) { // обработка нажатия Enter
		text := strings.TrimSpace(input.GetText()) // получаем текст из поля ввода и удаляем лишние пробелы
		if text == "" {
			return
		}
		input.SetText("")

		if text == "/exit" {
			app.Stop()
			return
		}

		var packet protocol.Packet

		if text == "/online" {
			packet = protocol.Packet{
				Type:    protocol.PacketMessage,
				From:    username,
				To:      "server",
				Payload: "/online",
			}
		} else if strings.HasPrefix(text, "@") {
			parts := strings.SplitN(text, " ", 2)
			if len(parts) < 2 {
				fmt.Fprintf(messages, "[red]Формат: @имя сообщение[-]\n")
				messages.ScrollToEnd()
				return
			}
			to := strings.TrimPrefix(parts[0], "@")
			packet = protocol.Packet{
				Type:    protocol.PacketMessage,
				From:    username,
				To:      to,
				Payload: parts[1],
			}
			// Показываем своё сообщение сразу с отметкой времени
			ts := time.Now().Format("15:04")
			fmt.Fprintf(messages, "[gray]%s[-] [blue]%s[-]: %s\n", ts, username, parts[1])
			messages.ScrollToEnd()
		} else {
			fmt.Fprintf(messages, "[red]Используйте @имя для отправки[-]\n")
			messages.ScrollToEnd()
			return
		}

		if err := protocol.Send(conn, packet); err != nil {
			fmt.Fprintf(messages, "[red]Ошибка отправки: %v[-]\n", err)
			messages.ScrollToEnd()
		}
	})

	// Горячие клавиши: F1 — запрос онлайн и фокус на списке; Tab — переключение фокуса
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			// Запрашиваем online и переключаем фокус на список
			protocol.Send(conn, protocol.Packet{
				Type:    protocol.PacketMessage,
				From:    username,
				To:      "server",
				Payload: "/online",
			})
			app.SetFocus(onlineList)
			return nil
		case tcell.KeyTAB:
			// Переключение между полем ввода и списком онлайн
			if app.GetFocus() == input {
				app.SetFocus(onlineList)
			} else {
				app.SetFocus(input)
			}
			return nil
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(layout, true).SetFocus(input).Run(); err != nil { 
		log.Fatalf("ошибка TUI: %v", err)
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