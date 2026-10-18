# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Что это

Telegram-бот (`github.com/askemblerrr/pr-review-bot`, Go 1.26), который принимает GitHub-вебхуки и шлёт уведомления о PR в Telegram. Пользовательские тексты бота, комментарии в коде и README — на русском; логи и ошибки — на английском.

Это учебный итоговый проект курса; требования — в `docs/tz.md`.

## Формат работы: менторство

Автор — новичок в Go и backend, цель — понять, а не просто получить код.

- Рутину (boilerplate, конфиги, Docker) пишет Claude; интерфейсы, бизнес-логику, обработку ошибок и ключевые тесты — автор, через `TODO(human)`.
- Перед каждым этапом — короткое обсуждение дизайна и «почему так», без кода.
- На просьбу о помощи с `TODO(human)` сначала давать намёк, готовое решение — только по прямой просьбе.
- Backend-термины пояснять при первом упоминании; ответы держать компактными.

## Команды

```bash
make run          # go run ./cmd/bot
make build        # go build -o bin/bot ./cmd/bot
make test         # go test -cover ./...
make lint         # golangci-lint run ./...  (конфига .golangci.* в репозитории нет — дефолтные линтеры)
make docker-up    # docker compose up --build -d + tail логов app
make docker-down
```

Один тест / один пакет:

```bash
go test ./internal/repository/memory -run TestUserRepositorySaveOverwrites -v
```

`make` на Windows может отсутствовать — тогда запускать команды `go` напрямую.

## Запуск локально

- Конфигурация только через переменные окружения (`internal/config`), `.env` сам приложением не читается — его подхватывает `docker compose`. При `make run` переменные нужно экспортировать вручную.
- Обязательные: `TELEGRAM_BOT_TOKEN`, `GITHUB_SECRET`, `GITHUB_USERNAME`. Необязательные: `PORT` (в формате адреса, `:8080`), `LOG_LEVEL`.
- База: `DB_HOST`, `DB_PORT` (5432), `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`. Пустой `DB_HOST` → in-memory хранилище с предупреждением в логе; если `DB_HOST` задан, логин и пароль обязательны. В compose `DB_HOST=postgres` проставляется сам, схема создаётся из `migrations/` через `docker-entrypoint-initdb.d` — только на пустом томе `pgdata` (после изменения схемы: `docker compose down -v`).
- HTTP-сервер поднимается только по TLS (`ListenAndServeTLS`), поэтому `TLS_CERT_PATH` / `TLS_KEY_PATH` обязательны (проверяются в `config.Load`). Сертификаты лежат в `certs/` (в `.gitignore`), в контейнер монтируются как `/app/certs`.
- Вебхук GitHub: `POST /api/v1/github/webhook`, наружу пробрасывается через `ngrok http https://localhost:8080` (сервер принимает только TLS 1.3, см. `newTLSConfig` в `cmd/bot/main.go`).

## Архитектура

Слоистая структура, всё связывается вручную в `cmd/bot/main.go` (`run`): repository → clients → services → delivery. Зависимости направлены внутрь: delivery → service → repository; между слоями ходят только доменные типы из `internal/types`.

Два входящих потока, работающих в отдельных горутинах:

1. **Telegram (long polling)**: `client/telegram.Client.Updates` владеет `GetUpdatesChan` и через `mapper.ToDomain` превращает `tgbotapi.Update` в `types.Message` (нетекстовые апдейты отбрасываются) → `delivery/telegram.Handler.Handle` → `router.go` (switch по `msg.Command`) → `service.UserService`. Типы `tgbotapi` не должны выходить за пределы `internal/client/telegram`.
2. **GitHub (webhook)**: `delivery/github.Handler.ServeHTTP` читает тело (лимит 1 МБ), проверяет HMAC `X-Hub-Signature-256` (`signature.go`), затем `router.go` диспетчеризует по `X-GitHub-Event`. Каждый тип события — отдельный файл (`pull_request.go`, `pull_request_review.go`, `pull_request_review_comment.go`) со своей структурой payload и фильтрацией, который маппит payload в `types.*Event` и вызывает `service.NotificationService`.

Важные особенности поведения:

- Фильтрация событий идёт в delivery-слое по единственному `GITHUB_USERNAME` из конфига (назначен на меня / мой PR / не мой собственный комментарий). То есть бот по сути однопользовательский, хотя репозиторий хранит маппинг telegramID ↔ githubLogin для многих.
- Вебхук всегда отвечает `200`, даже если обработка события упала (ошибка только логируется) — чтобы GitHub не ретраил. `ErrUserNotFound` из `NotificationService` тоже попадает в этот лог как error.
- Хранилище выбирается в `main.go`: `repository/postgres` (pgxpool) или `repository/memory`. Любой бэкенд реализует `repository.UserRepository`, хранит логин в нижнем регистре (и приводит его при поиске), возвращает `repository.ErrUserNotFound` при отсутствии записи и `repository.ErrLoginTaken`, если логин привязан к другому telegramID (в Postgres — по коду `23505` от `UNIQUE(github_login)`). Delivery видит эту ошибку как `service.ErrLoginTaken`.
- Исходящая отправка абстрагирована интерфейсами, объявленными у потребителя: `service.Notifier` и `delivery/telegram.Sender` (оба реализует `client/telegram.Client`).

Добавление нового события GitHub: константа + `case` в `delivery/github/router.go`, файл-обработчик с payload-структурой, тип в `internal/types/event.go`, метод `Notify*` в `NotificationService`. Новая команда Telegram: `case` в `delivery/telegram/router.go` + метод-обработчик в `handler.go`.

## Текущее состояние (проверено `go vet` / `go test`)

- Тесты сервисов — заготовки: `user_test.go` в пакете `service`, а `mocks_test.go` — в `service_test`, так что моки оттуда не видны; `mockUserRepo.GetGitHubLogin` принимает `string` вместо `int64` и не удовлетворяет `repository.UserRepository`. В `mocks_test.go` начат переход на `testify/mock`.
- `internal/validators` нигде не используется; `pkg/` пуст.
- Правила фильтрации расходятся: `pull_request_review_thread` фильтруется по `assignees` (как в ТЗ, без учёта регистра), ревью и комментарии — по автору PR через `==`.
- У `repository/postgres` нет постоянных тестов (проверялся вручную на контейнере).