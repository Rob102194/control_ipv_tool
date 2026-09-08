// Command desktop es el ejecutable de escritorio (Wails v2), multiplataforma
// (Windows, macOS, Linux).
//
// No lleva lógica de negocio: arma la aplicación con appboot (el mismo cableado
// que cmd/server) y usa el router como AssetServer.Handler de Wails, de modo que
// el frontend y la API viajan por el mismo http.Handler. Así, migrar a web solo
// implica compilar cmd/server en vez de este binario.
//
// Compilar este paquete arrastra el toolchain de Wails (CGO + WebKit/GTK en
// Linux). El CI instala esas dependencias; en macOS y Windows no hacen falta
// paquetes extra.
package main

import (
	"context"

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

	var ctxRef context.Context

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
				if ctxRef != nil {
					wruntime.WindowUnminimise(ctxRef)
					wruntime.WindowShow(ctxRef)
				}
			},
		},
		OnStartup: func(ctx context.Context) {
			ctxRef = ctx
			logger.Info("escritorio listo", "db", app.DBPath)
		},
		OnShutdown: func(context.Context) {
			_ = app.Close()
		},
	})
	if err != nil {
		logger.Error("Wails terminó con error", "err", err)
	}
}
