# Конфигурация проекта

Проект настраивается через `.env` в корне репозитория.

## Основные переменные

```env
# Database
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=DiplomEnglish
DB_PORT=5432

# API
API_PORT=9090

# Frontend
FRONTEND_PORT=3000
FORCE_HTTPS_REDIRECT=false
ALLOW_IFRAME_EMBED=true
IFRAME_ANCESTORS=

# Adminer
ADMINER_PORT=8080

# JWT
JWT_SECRET=change_this_secret
JWT_EXPIRY_HOURS=24

# App
GIN_MODE=release

# Moodle
MOODLE_ENABLED=true
MOODLE_TEST_MODE=false
MOODLE_BASE_URL=https://your-moodle.example.com
MOODLE_TOKEN=your_moodle_web_service_token
MOODLE_SERVICE=moodle_mobile_app
MOODLE_AUTO_CREATE=true
MOODLE_ONLY_AUTH=true
```

## Что публикует Docker Compose

- `backend` публикуется на `API_PORT`
- `frontend` публикуется на `FRONTEND_PORT`
- `adminer` публикуется на `ADMINER_PORT`
- `db` публикуется на `DB_PORT`

При этом пользователь должен заходить именно на `frontend`, потому что:

- он отдает интерфейс
- он же проксирует `/api` и `/uploads` в backend

## HTTPS-ready режим

Для продакшена лучше использовать такую схему:

1. Домен указывает на внешний reverse proxy или хостинг-панель
2. HTTPS завершается там
3. Прокси передает трафик в контейнер `frontend`
4. `frontend` уже отправляет `/api` и `/uploads` в `backend`

`FORCE_HTTPS_REDIRECT`:

- `false` для локальной разработки по `http://localhost`
- `true` для продакшена, когда перед приложением уже настроен TLS и заголовок `X-Forwarded-Proto`

`ALLOW_IFRAME_EMBED` и `IFRAME_ANCESTORS`:

- `ALLOW_IFRAME_EMBED=true` разрешает открывать приложение внутри `iframe`
- пустой `IFRAME_ANCESTORS` снимает ограничения полностью
- если нужно ограничить родительский сайт, укажите, например, `IFRAME_ANCESTORS=https://portal.example.com`

## Что важно для Moodle-only входа

- при `MOODLE_ONLY_AUTH=true` локальные `/api/auth/login` и `/api/auth/register` возвращают отказ
- основной вход выполняется через `/api/auth/login/moodle`
- UI страницы авторизации теперь соответствует этой логике и не предлагает локальную регистрацию

## После изменения конфигурации

```bash
docker compose down
docker compose up --build -d
```
