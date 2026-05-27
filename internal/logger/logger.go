package logger

import (
    "fmt"
    "io"
    "log"
    "os"
    "time"
    "strings"
)

const (
	LevelDebug = iota // 0
	LevelInfo         // 1
	LevelWarn         // 2
	LevelError        // 3
)

type Logger struct {
    file *os.File      
    logger *log.Logger 
    logLevel int
}

func NewLogger(filename string, levelStr string) (*Logger, error) {

    var lvl int
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		lvl = LevelDebug
	case "info":
		lvl = LevelInfo
	case "warn":
		lvl = LevelWarn
	case "error":
		lvl = LevelError
	default:
		// Info по умолчанию
		lvl = LevelInfo
	}

    
    // Создаем директорию для логов, если она не существует
    // 0755 - права доступа к директории (чтение и выполнение для всех, запись для владельца)
    if err := os.MkdirAll("logs", 0755); err != nil {
        return nil, fmt.Errorf("не удалось создать директорию для логов: %w", err)
    }
    
    filePath := fmt.Sprintf("logs/%s_%s", filename, time.Now().Format("02-01-2006"))

    // Открываем файл для логирования (создаем, если не существует, и добавляем в конец)
    // 0644 - права доступа к файлу (чтение и запись для владельца, чтение для остальных)
    file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, fmt.Errorf("не удалось открыть файл для логирования: %w", err)
    }

    // Логируем и в файл, и в косоль
    multiWriter := io.MultiWriter(os.Stdout, file)

    return &Logger{
        file: file,
        logger: log.New(multiWriter, "", log.LstdFlags), 
        logLevel: lvl,
    }, nil
}
func (l *Logger) isLogging(msgLevel int) bool {
	return msgLevel >= l.logLevel
}

func (l *Logger) Close() {
    l.file.Close()
}

func (l *Logger) Debug(format string, v ...interface{}) {
    if !l.isLogging(LevelDebug) {
		return
	}
    l.logger.Printf("[Debug] " + format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
    if !l.isLogging(LevelInfo) {
		return
	}
    l.logger.Printf("[INFO] " + format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
    if !l.isLogging(LevelError) {
		return
	}
    l.logger.Printf("[ERROR] " + format, v...)
}

func (l *Logger) Auth(username, event string) { 
    l.logger.Printf("[AUTH]  %s: %s", username, event) 
}

func (l *Logger) Msg(from, to, text string) { 
    l.logger.Printf("[MSG]   %s -> %s: %s", from, to, text) 
}