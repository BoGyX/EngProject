# Backend - English Learning Platform

## Шаг 1: Подключение к БД и Swagger

Это минимальная версия для начала работы.

## Требования

- Go 1.21+
- PostgreSQL 14+

## Настройка

### 1. Создайте базу данных

```bash
psql -U postgres
CREATE DATABASE english_learning;
\q
```

### 2. Примените схему БД

```bash
psql -U postgres -d english_learning -f ../schema.sql
```

### 3. Создайте .env файл

```bash
cp .env.example .env
```

Отредактируйте `.env` и укажите свои данные:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=ваш_пароль
DB_NAME=english_learning
```

### 4. Установите зависимости

```bash
go mod download
```

### 5. Установите Swagger CLI (для генерации документации)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 6. Сгенерируйте Swagger документацию

```bash
swag init -g main.go -o ./docs
```

### 7. Запустите сервер

```bash
go run main.go
```

## Проверка

После запуска:

1. **Health check**: http://localhost:8000/health
   - Должен показать статус сервера и БД

2. **Swagger UI**: http://localhost:8000/swagger/index.html
   - Интерактивная документация API

3. **API info**: http://localhost:8000/api
   - Информация об API

## Структура проекта

```
backend/
├── main.go                  # Точка входа
├── internal/
│   ├── config/
│   │   └── config.go        # Конфигурация из .env
│   └── database/
│       └── database.go      # Подключение к PostgreSQL
├── docs/                    # Сгенерированная Swagger документация (после swag init)
├── .env                     # Ваши настройки (не в git)
├── .env.example             # Пример настроек
└── go.mod                   # Зависимости
```

## Что дальше?

После успешного запуска:
1. ✅ Проверьте подключение к БД через `/health`
2. ✅ Откройте Swagger документацию
3. 📝 Следующий шаг: добавим первый endpoint для работы со словами
