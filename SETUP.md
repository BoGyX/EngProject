# Настройка проекта

Все основные параметры задаются через `.env` в корне проекта.

## Быстрый запуск

```bash
docker compose up --build -d
```

Эта команда использует production-конфигурацию:

- `backend` запускается как собранный бинарник
- `frontend` запускается как production-сервер на Node.js
- dev-зависимости вроде `go run`, `vite` и `air` для сервера не нужны

## Режим разработки без перезапуска контейнеров

Для локальной разработки с bind mount:

```bash
docker compose -f docker-compose.dev.yml up --build
```

Что это дает:

- `backend` запускается обычной командой `go run .`
- `frontend` запускается через `vite` и подхватывает изменения без ручного рестарта
- `uploads` и исходники монтируются из рабочей папки, поэтому изменения видны сразу

Если раньше уже поднимали старую версию compose и видите ошибку авторизации Postgres, текущая конфигурация использует отдельный volume `postgres_data_v2`, чтобы база инициализировалась заново с актуальными значениями из `.env`.

## Доступ к сервисам

После запуска:

- Frontend: `http://localhost:3000` или значение `FRONTEND_PORT`
- API: `http://localhost:9090` или значение `API_PORT`
- Swagger: `http://localhost:9090/swagger/index.html`
- Adminer: `http://localhost:8080` или значение `ADMINER_PORT`
- PostgreSQL: `localhost:5432` или значение `DB_PORT`

## Moodle-only авторизация

Проект теперь рассчитан на вход только через Moodle:

- страница `/login` принимает только `username` и `password` Moodle
- локальная регистрация и обычный login отключаются при `MOODLE_ONLY_AUTH=true`
- при первом успешном входе локальный профиль создается автоматически, если включен `MOODLE_AUTO_CREATE=true`

Минимальный набор переменных:

```env
MOODLE_ENABLED=true
MOODLE_TEST_MODE=false
MOODLE_BASE_URL=https://your-moodle.example.com
MOODLE_TOKEN=your_moodle_web_service_token
MOODLE_SERVICE=moodle_mobile_app
MOODLE_AUTO_CREATE=true
MOODLE_ONLY_AUTH=true
```

## HTTPS и хостинг

В проекте больше нет отдельного `nginx`. Публичной точкой входа является контейнер `frontend`.

Как это устроено:

- frontend сам отдает SPA
- frontend проксирует `/api` и `/uploads` в backend
- поэтому для HTTPS достаточно поставить внешний reverse proxy или панель хостинга перед `frontend`

Рекомендуемая схема:

1. Поднять `docker compose up -d`
2. Опубликовать наружу только `FRONTEND_PORT`
3. Настроить HTTPS на домене до контейнера `frontend`
4. После включения TLS при необходимости установить `FORCE_HTTPS_REDIRECT=true`

Важно:

- `FORCE_HTTPS_REDIRECT=true` включайте только когда внешний прокси уже передает `X-Forwarded-Proto=https`
- фронтенд использует относительные пути `/api` и `/uploads`, поэтому mixed content при нормальном reverse proxy не будет
- загруженные аудио и изображения сохраняются как относительные пути, поэтому домен и протокол можно менять без поломки карточек
- для LMS по умолчанию разрешён домен `https://lms.25-12.ru`; при необходимости его можно заменить через `IFRAME_ANCESTORS`

## Если меняли `.env`

Перезапуск:

```bash
docker compose down
docker compose up --build -d
```
