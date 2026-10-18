# Pull Request Reviews Bot

Telegram-бот, уведомляющий о назначении PR на вас в GitHub.

## Запуск

1. Скопируйте `.env.example` в `.env` и заполните (включая `DB_USERNAME` / `DB_PASSWORD` для PostgreSQL).
2. Сгенерируйте сертификаты: `openssl req ...`
3. `make docker-up`
4. `ngrok http https://localhost:8080` и настройте webhook в GitHub.