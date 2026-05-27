package logger

import (
    "fmt"
    "io"
    "log"
    "os"
    "time"
)

type Logger struct {
    file *os.File      
    logger *log.Logger 
}

func NewLogger(filename string) (*Logger, error) {
    
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
    }, nil
}

func (l *Logger) Close() {
    l.file.Close()
}

func (l *Logger) Info(format string, v ...interface{}) {
    l.logger.Printf("[INFO] " + format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
    l.logger.Printf("[ERROR] " + format, v...)
}

func (l *Logger) Auth(username, event string) { 
    l.logger.Printf("[AUTH]  %s: %s", username, event) 
}

func (l *Logger) Msg(from, to, text string) { 
    l.logger.Printf("[MSG]   %s -> %s: %s", from, to, text) 
}