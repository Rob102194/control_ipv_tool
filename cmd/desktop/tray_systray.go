//go:build windows || linux

package main

import (
	_ "embed"

	"github.com/energye/systray"
)

//go:embed icon.png
var trayIcon []byte

// startTray monta el icono de la bandeja (Windows/Linux). Best-effort: si falla,
// la app sigue funcionando (cerrar la ventana la oculta y se reabre por el SO).
func startTray(sh *shell) {
	defer func() {
		if r := recover(); r != nil {
			sh.logger.Warn("no se pudo iniciar la bandeja del sistema", "err", r)
		}
	}()
	systray.Run(func() {
		systray.SetTitle("Control IPV")
		systray.SetTooltip("Control IPV")
		if len(trayIcon) > 0 {
			systray.SetIcon(trayIcon)
		}
		abrir := systray.AddMenuItem("Abrir", "Mostrar la ventana")
		systray.AddSeparator()
		salir := systray.AddMenuItem("Salir", "Cerrar Control IPV")

		abrir.Click(func() { sh.show() })
		salir.Click(func() { sh.quit() })
		systray.SetOnClick(func(systray.IMenu) { sh.show() })
	}, nil)
}

func stopTray() { systray.Quit() }
