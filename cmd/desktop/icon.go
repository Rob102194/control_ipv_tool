package main

import _ "embed"

// trayIcon es el icono de la bandeja del sistema (PNG en Windows/Linux;
// en macOS systray lo adapta).
//
//go:embed icon.png
var trayIcon []byte
