// Command server arranca la API de Control IPV como servidor HTTP independiente.
//
// Es la entrega usada para desarrollo del frontend y, más adelante, para el
// despliegue web. El binario de escritorio (cmd/desktop) reutiliza el mismo
// appboot; aquí solo cambia el envoltorio.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rob102194/control_ipv_tool/internal/appboot"
	"github.com/Rob102194/control_ipv_tool/internal/platform"
	"github.com/Rob102194/control_ipv_tool/web"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := platform.LoadConfig()
	if err != nil {
		println("config:", err.Error())
		return err
	}
	logger := platform.SetupLogging(cfg.LogLevel, cfg.Env)

	app, err := appboot.New(cfg, logger, web.Handler())
	if err != nil {
		logger.Error("no se pudo iniciar la aplicación", "err", err)
		return err
	}
	defer app.Close()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return serve(srv, logger, app.DBPath)
}

// serve arranca el servidor y lo apaga con gracia ante SIGINT/SIGTERM.
func serve(srv *http.Server, logger *slog.Logger, dbPath string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("servidor escuchando", "addr", srv.Addr, "db", dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		logger.Error("el servidor terminó de forma inesperada", "err", err)
		return err
	case <-ctx.Done():
		logger.Info("apagando servidor…")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("apagado forzado", "err", err)
			return err
		}
		logger.Info("servidor detenido")
		return nil
	}
}
