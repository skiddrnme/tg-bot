package main

import (
    "context"
    "crypto/tls"
    "errors"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "github.com/askemblerrr/pr-review-bot/internal/client/telegram"
    "github.com/askemblerrr/pr-review-bot/internal/config"
    ghdelivery "github.com/askemblerrr/pr-review-bot/internal/delivery/github"
    tgdelivery "github.com/askemblerrr/pr-review-bot/internal/delivery/telegram"
    "github.com/askemblerrr/pr-review-bot/internal/logger"
    "github.com/askemblerrr/pr-review-bot/internal/repository"
    "github.com/askemblerrr/pr-review-bot/internal/repository/memory"
    "github.com/askemblerrr/pr-review-bot/internal/repository/postgres"
    "github.com/askemblerrr/pr-review-bot/internal/service"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    if err := run(ctx); err != nil {
        fmt.Fprintln(os.Stderr, "fatal:", err)
        os.Exit(1)
    }
}

func run(ctx context.Context) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("config: %w", err)
    }
    log := logger.New(cfg.LogLevel)
    log.Info("starting pr-review-bot")

    botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
    if err != nil {
        return fmt.Errorf("telegram bot: %w", err)
    }
    log.Info("telegram bot authorized", "username", botAPI.Self.UserName)

    // Repository
    var userRepo repository.UserRepository
    if cfg.DB.Enabled() {
        connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
        pgRepo, err := postgres.NewUserRepository(connectCtx, cfg.DB.DSN())
        cancel()
        if err != nil {
            return fmt.Errorf("postgres: %w", err)
        }
        defer pgRepo.Close()
        userRepo = pgRepo
        log.Info("using postgres storage", "host", cfg.DB.Host, "db", cfg.DB.Name)
    } else {
        userRepo = memory.NewUserRepository()
        log.Warn("DB_HOST is not set, using in-memory storage: data is lost on restart")
    }

    // Clients
    tgClient := telegram.NewBot(botAPI, log)

    // Services
    userSvc := service.NewUserService(userRepo, log)
    notifSvc := service.NewNotificationService(userRepo, tgClient, log)

    // Delivery
    tgHandler := tgdelivery.NewHandler(userSvc, tgClient, log)
    ghHandler := ghdelivery.NewHandler(notifSvc, cfg.GithubUsername, cfg.GithubSecret, log)

    // HTTP-сервер для GitHub
    mux := http.NewServeMux()
    mux.Handle("POST /api/v1/github/webhook", ghHandler)

    srv := newServer(cfg.Port, mux)

    errCh := make(chan error, 2)

    // HTTP-сервер
    go func() {
        log.Info("http server started", "addr", cfg.Port)
        if err := srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath); err != nil &&
            !errors.Is(err, http.ErrServerClosed) {
            errCh <- fmt.Errorf("http server: %w", err)
        }
    }()

    // Telegram polling
    go func() {
        log.Info("telegram polling started")
        for msg := range tgClient.Updates(ctx) {
            tgHandler.Handle(ctx, msg)
        }
        log.Info("telegram polling stopped")
    }()

    select {
    case <-ctx.Done():
        log.Info("shutdown signal received")
    case err := <-errCh:
        log.Error("background error", "err", err)
    }

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        return fmt.Errorf("shutdown: %w", err)
    }
    log.Info("shutdown complete")
    return nil
}

func newServer(addr string, handler http.Handler) *http.Server {
    return &http.Server{
        Addr:              addr,
        Handler:           handler,
        TLSConfig:         newTLSConfig(),
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
        MaxHeaderBytes:    1 << 16,
    }
}

func newTLSConfig() *tls.Config {
    return &tls.Config{MinVersion: tls.VersionTLS13}
}