package main

import (
	"bufio" // буфферизированный ввод-вывод
	"fmt"   // форматированный вывод
	"log"   // логирование
	"net"   // сетевые операции
	"os"    // работа с оc
	"strings" // работа со строками
	"time"	// работа со временем
	"flag"	// парсинг флагов командной строки


	"github.com/fulikozzz/AuroraMsgr/internal/protocol"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"

)

var (
	serverAddr = flag.String("server", "127.0.0.1:8080", "адрес сервера в формате host:port")
)

func main(){
	flag.Parse()

	fmt.Printf("----- Aurora -----\n")
	fmt.Printf("подключение к %s...\n", *serverAddr)

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close() // закрываем соединение при завершении рабоы

	log.Printf("подключено к серверу по адресу %s", *serverAddr)

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

func runInterface(conn net.Conn, username string) {
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

	// Список диалогов
	dialogList := tview.NewList()
	dialogList.SetBorder(true)
	dialogList.SetTitle(" Диалоги ")

	// Список онлайн пользователей
	onlineList := tview.NewList()
	onlineList.SetBorder(true)
	onlineList.SetTitle(" Онлайн ")

	// Текущий собеседник
	currentChat := ""

	// Загрузка истории диалога
	loadHistory := func(partner string) {
		currentChat = partner
		messages.Clear()
		messages.SetTitle(fmt.Sprintf(" Диалог с %s ", partner))
		protocol.Send(conn, protocol.Packet{
			Type:    protocol.PacketHistoryRequest,
			From:    username,
			Payload: partner,
		})
		input.SetText("@" + partner + " ")
		app.SetFocus(input)
	}

	// Выбор диалога
	dialogList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		loadHistory(strings.TrimSpace(mainText))
	})

	// Выбор из онлайн списка
	onlineList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		loadHistory(strings.TrimSpace(mainText))
	})

	// Подсказки
	hints := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	hints.SetBorder(true).SetTitle(" Подсказки ")
	hints.SetText(`[yellow]Команды:[-]
@имя сообщение — написать
/online — обновить онлайн
/exit   — выход
[green]Клавиши:[-]
F1 — фокус онлайн
F2 — фокус диалоги
Tab — поле ввода`)

	// Правый столбец
	rightFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dialogList, 0, 1, false).
		AddItem(onlineList, 0, 1, false).
		AddItem(hints, 10, 0, false)

	// Основной layout
	contentFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(messages, 0, 1, false).
		AddItem(rightFlex, 30, 0, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, false).
		AddItem(input, 3, 0, true)

	// Запрашиваем диалоги и онлайн при входе
	protocol.Send(conn, protocol.Packet{
		Type:    protocol.PacketDialogsRequest,
		From:    username,
		Payload: username,
	})
	protocol.Send(conn, protocol.Packet{
		Type:    protocol.PacketMessage,
		From:    username,
		To:      "server",
		Payload: "/online",
	})

	// Автообновление онлайн каждые 10 секунд
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

	// Горутина получения сообщений
	go func() {
		for {
			packet, err := protocol.Receive(conn)
			if err != nil {
				fmt.Fprintf(messages, "[red]соединение потеряно[-]\n")
				app.Draw()
				return
			}

			switch packet.Type {
			case protocol.PacketMessage:
				ts := time.Now().Format("15:04")
				fmt.Fprintf(messages, "[gray]%s[-] [green]%s[-]: %s\n", ts, packet.From, packet.Payload)
				messages.ScrollToEnd()

				// Обновляем диалоги при новом сообщении
				protocol.Send(conn, protocol.Packet{
					Type:    protocol.PacketDialogsRequest,
					From:    username,
					Payload: username,
				})

			case protocol.PacketHistory:
				messages.Clear()
				if packet.Payload != "" {
					fmt.Fprintf(messages, "%s\n", packet.Payload)
				} else {
					fmt.Fprintf(messages, "[gray]История пуста[-]\n")
				}
				messages.ScrollToEnd()

			case protocol.PacketDialogs:
				dialogList.Clear()
				if packet.Payload != "" {
					for _, d := range strings.Split(packet.Payload, ",") {
						d = strings.TrimSpace(d)
						if d != "" {
							dialogList.AddItem(d, "", 0, nil)
						}
					}
				}
				app.Draw()

			case protocol.PacketSystem:
				const prefix = "Пользователи онлайн:"
				if strings.HasPrefix(packet.Payload, prefix) {
					list := strings.TrimSpace(strings.TrimPrefix(packet.Payload, prefix))
					onlineList.Clear()
					if list != "" {
						for _, u := range strings.Split(list, ",") {
							u = strings.TrimSpace(u)
							if u != "" {
								onlineList.AddItem(u, "", 0, nil)
							}
						}
					}
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

	// Обработка ввода
	input.SetDoneFunc(func(key tcell.Key) {
		text := strings.TrimSpace(input.GetText())
		if text == "" {
			return
		}
		input.SetText("")

		defer func() {
			if currentChat != "" {
				input.SetText("@" + currentChat + " ")
			}
		}()

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
			ts := time.Now().Format("15:04")
			fmt.Fprintf(messages, "[gray]%s[-] [blue]%s[-]: %s\n", ts, username, parts[1])
			messages.ScrollToEnd()

			// Если открыт диалог — обновляем имя получателя
			if currentChat == "" {
				currentChat = to
				messages.SetTitle(fmt.Sprintf(" Диалог с %s ", to))
			}
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

	// Горячие клавиши
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			app.SetFocus(onlineList)
			return nil
		case tcell.KeyF2:
			app.SetFocus(dialogList)
			return nil
		case tcell.KeyTAB:
			app.SetFocus(input)
			return nil
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyPgUp:
			row, _ := messages.GetScrollOffset()
			messages.ScrollTo(row-5, 0)
			return nil
		case tcell.KeyPgDn:
			row, _ := messages.GetScrollOffset()
			messages.ScrollTo(row+5, 0)
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