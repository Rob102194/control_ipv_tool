//go:build !windows && !linux

package main

// En macOS no se monta bandeja: la librería systray toma el runloop principal
// de Cocoa, que ya lo gestiona Wails, y el enlazado falla. Cerrar la ventana la
// oculta igualmente (OnBeforeClose) y se reabre desde el Dock.
func startTray(sh *shell) {
	sh.logger.Info("bandeja del sistema no disponible en esta plataforma; cerrar la ventana la oculta")
}

func stopTray() {}
