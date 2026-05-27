# AuroraMsgr
Консольный мессенджер (клиент+сервер) на Go. Курсовая работа по Операционным системам

# Stack
- Язык: Go
- БД: SQLite
- Сетевой протокол: TCP

## Architecture
```path
cmd/ 
    client/ - код клиента
    server/ - код сервера
internal/
    auth/ - логика аутентификации
    chat/ - логика обмена сообщениями
    logger/ - логирование
    protocol/ - cетевой протокол
    storage/ - БД
config/ - конфигурационные файлы
docs/ - документация
logs/ - файлы логов
```

## Build

### Server
```bash
go build -o aurora-server ./cmd/server
go run ./cmd/server
```

### Client
```bash
go build -o aurora-client ./cmd/client
go run ./cmd/client
```

