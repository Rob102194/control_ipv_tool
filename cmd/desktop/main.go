// Command desktop es el ejecutable de escritorio (Wails v2).
//
// No lleva lógica de negocio: arma la aplicación con appboot (el mismo cableado
// que cmd/server) y usa el router como AssetServer.Handler de Wails, de modo que
// el frontend y la API viajan por el mismo http.Handler. Así, migrar a web solo
// implica compilar cmd/server en vez de este binario.
//
// Compilar este paquete arrastra el toolchain de Wails (CGO + WebKit/GTK en
// Linux). El CI instala esas dependencias; en macOS y Windows no hacen falta
// paquetes extra. La bandeja del sistema (systray) solo se compila en Windows y
// Linux; en macOS entra en conflicto con el runloop de Cocoa de Wails (ver
// tray_noop.go).
package main

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/Rob102194/control_ipv_tool/internal/appboot"
	"github.com/Rob102194/control_ipv_tool/internal/platform"
	"github.com/Rob102194/control_ipv_tool/web"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		panic(err)
	}
	logger := platform.SetupLogging(cfg.LogLevel, cfg.Env)

	app, err := appboot.New(cfg, logger, web.Handler())
	if err != nil {
		logger.Error("no se pudo iniciar la aplicación", "err", err)
		return
	}

	sh := &shell{logger: logger}

	err = wails.Run(&options.App{
		Title:     "Control IPV",
		Width:     1280,
		Height:    850,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			// El router (SPA embebido + /api) es el asset server de Wails.
			Handler: app.Handler,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "com.control-ipv.desktop",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				sh.show()
			},
		},
		OnStartup: func(ctx context.Context) {
			sh.ctx = ctx
			logger.Info("escritorio listo", "db", app.DBPath)
			go startTray(sh)
		},
		OnBeforeClose: func(context.Context) (prevent bool) {
			// Cerrar la ventana la oculta a la bandeja; se sale desde el menú
			// de la bandeja ("Salir"). En macOS (sin bandeja) simplemente
			// oculta la ventana; se puede reabrir desde el Dock.
			if sh.quitting.Load() {
				return false
			}
			sh.hide()
			return true
		},
		OnShutdown: func(context.Context) {
			stopTray()
			_ = app.Close()
		},
	})
	if err != nil {
		logger.Error("Wails terminó con error", "err", err)
	}
}

// shell mantiene el contexto de Wails y coordina ventana + bandeja.
type shell struct {
	ctx      context.Context
	logger   *slog.Logger
	quitting atomic.Bool
}

func (s *shell) show() {
	if s.ctx != nil {
		wruntime.WindowShow(s.ctx)
	}
}

func (s *shell) hide() {
	if s.ctx != nil {
		wruntime.WindowHide(s.ctx)
	}
}

func (s *shell) quit() {
	s.quitting.Store(true)
	if s.ctx != nil {
		wruntime.Quit(s.ctx)
	}
}
