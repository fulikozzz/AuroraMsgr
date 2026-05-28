.PHONY: build build-linux build-windows clean

# Сборка для текущей платформы
build:
	go build -o aurora-server ./cmd/server
	go build -o aurora-client ./cmd/client

# Сборка для Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o build/linux/aurora-server ./cmd/server
	GOOS=linux GOARCH=amd64 go build -o build/linux/aurora-client ./cmd/client

# Сборка для Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o build/windows/aurora-server.exe ./cmd/server
	GOOS=windows GOARCH=amd64 go build -o build/windows/aurora-client.exe ./cmd/client

# Сборка для всех платформ
build-all: build-linux build-windows

# Очистка
clean:
	rm -rf build/ aurora-server aurora-client aurora-server.exe aurora-client.exe