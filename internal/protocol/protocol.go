package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

// Типы пакетом
const (
	PacketMessage = "message"	// обычное сообщение
	PacketSystem  = "system"	// системное сообщение
	PacketLogin    = "login"	// запрос на вход
	PacketRegister = "register"	// запрос на регистрацию
	PacketSuccess  = "success"	// успешный ответ
	PacketError    = "error"	// ошибка
	PacketHistory	= "history"	// история сообщений
	PacketHistoryRequest = "history_request" // запрос на историю сообщений
	PacketDialogs	= "dialogs"	// список диалогов
	PacketDialogsRequest = "dialogs_request" // запрос на список диалогов

)

// Структура пакета
type Packet struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	To      string `json:"to"`
	Payload string `json:"payload"`
}


// Функция отправки пакета
func Send(conn net.Conn, packet Packet) error {
	data, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("не удалось собрать пакет: %w", err)
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}

// Функция получения пакета
func Receive(conn net.Conn) (Packet, error) {
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return Packet{}, fmt.Errorf("чтение пакета: %w", err)
	}
	data = bytes.TrimSpace(data)
	var p Packet
	if err := json.Unmarshal(data, &p); err != nil {
		return Packet{}, fmt.Errorf("не удалось разобрать пакет: %w", err)
	}
	return p, nil
}