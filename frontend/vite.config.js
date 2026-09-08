import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// El servidor Go de desarrollo (cmd/server) escucha en 127.0.0.1:5175.
const API_TARGET = process.env.VITE_DEV_API || 'http://127.0.0.1:5175'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  envDir: './env',
  build: {
    // El build va a web/dist para que Go lo embeba (internal → web/embed.go).
    outDir: '../web/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': { target: API_TARGET, changeOrigin: true },
    },
  },
})
