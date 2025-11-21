.PHONY: build run test lint clean docker-build docker-up docker-down help deps verify fmt

BINARY_NAME=assignment-service
DOCKER_COMPOSE=docker-compose

build: ## Собрать бинарник
	@echo "Сборка приложения..."
	go build -o bin/$(BINARY_NAME) ./cmd/server

run: ## Запустить приложение локально
	@echo "Запуск приложения..."
	go run ./cmd/server

test: ## Запустить тесты с race detector
	@echo "Запуск тестов..."
	go test -v -race ./...

test-coverage: test ## Показать покрытие тестами
	@echo "Покрытие кода:"
	go tool cover -func=coverage.out

lint: ## Проверить код golangci-lint
	@echo "Проверка кода линтером..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint не установлен. Установите его: https://golangci-lint.run/"; \
	fi

fmt: ## Отформатировать код (go fmt + goimports)
	@echo "Форматирование кода..."
	go fmt ./...
	goimports -w .

verify: fmt lint test  ## Полная проверка: формат + линтер + тесты
	@echo "Проверка завершена успешно!"

clean: ## Удалить бинарники и coverage
	@echo "Очистка..."
	rm -rf bin/
	rm -f coverage.out

deps: ## Скачать и почистить зависимости
	@echo "Установка зависимостей..."
	go mod download
	go mod tidy

docker-build: ## Собрать Docker-образы
	@echo "Сборка Docker образа..."
	$(DOCKER_COMPOSE) build

docker-up: ## Запустить сервисы
	@echo "Запуск сервисов..."
	$(DOCKER_COMPOSE) up

docker-up-detached: ## Запустить сервисы (в фоне)
	@echo "Запуск сервисов..."
	$(DOCKER_COMPOSE) up -d

docker-down: ## Остановить и удалить контейнеры
	@echo "Остановка сервисов..."
	$(DOCKER_COMPOSE) down

docker-clean: docker-down ## Полная очистка: контейнеры + volumes + образы
	@echo "Удаляем всё, включая базу данных..."
	docker compose down -v --remove-orphans
	docker volume rm assignment-service_mongodb_data 2>/dev/null || true
	@echo "Всё удалено. Теперь можно запускать чистую БД: make docker-up"

docker-logs: ## Следить за логами контейнеров
	$(DOCKER_COMPOSE) logs -f

help: ## Показать эту справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-15s\033[0m %s\n", $$1, $$2}'
