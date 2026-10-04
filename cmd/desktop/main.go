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
	"crypto/sha256"
	"encoding/hex"

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
		Title:     tituloVentana(app),
		Width:     1280,
		Height:    850,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			// El router (SPA embebido + /api) es el asset server de Wails.
			Handler: app.Handler,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			// El candado de instancia única es un mutex/lock a nivel de
			// sistema operativo (ver wails/v2 internal/frontend/desktop/*)
			// identificado por este UniqueId. Si fuera un string fijo, dos
			// negocios con bases de datos distintas (CONTROL_IPV_DATA_DIR)
			// no podrían tener cada uno su propia ventana abierta a la vez:
			// el segundo lanzamiento solo traería al frente al primero. Al
			// derivarlo de la ruta de la BD, cada negocio tiene su propio
			// candado (se puede abrir una vez cada uno en paralelo), pero
			// abrir el MISMO negocio dos veces sigue bloqueado como antes.
			UniqueId: "com.control-ipv.desktop." + hashCorto(app.DBPath),
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

// hashCorto reduce una ruta a un identificador corto y estable, apto para
// nombrar un mutex/lock de sistema operativo (sin separadores de ruta ni
// límite de longitud que preocupe).
func hashCorto(s string) string {
	suma := sha256.Sum256([]byte(s))
	return hex.EncodeToString(suma[:8])
}

// tituloVentana arma "Control IPV" o "Control IPV — <negocio>" según el
// ajuste guardado en la BD de este proceso. Importa sobre todo cuando hay
// varias instancias abiertas a la vez (una por negocio, ver
// SingleInstanceLock más abajo): sin esto, Windows/macOS muestran "Control
// IPV" en la barra de tareas y el selector de ventanas para todas, sin forma
// de distinguir cuál es cuál antes de hacer clic. Se lee una sola vez al
// arrancar; si el usuario cambia el nombre desde la app, el título de la
// ventana se actualiza recién al reabrirla.
func tituloVentana(app *appboot.App) string {
	const base = "Control IPV"
	if app.Services == nil {
		return base
	}
	cfg, err := app.Services.Configuracion.Obtener(context.Background())
	if err != nil || cfg.NombreNegocio == "" {
		return base
	}
	return base + " — " + cfg.NombreNegocio
}
