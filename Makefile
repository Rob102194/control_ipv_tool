# Control IPV — Go + Wails.
# Backend Go en cmd/ + internal/; frontend React en frontend/.

GO        ?= go
BIN_DIR   ?= bin
SERVER_BIN := $(BIN_DIR)/control-ipv-server

.DEFAULT_GOAL := help

.PHONY: help
help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Ordena go.mod / go.sum
	$(GO) mod tidy

.PHONY: frontend
frontend: ## Compila el SPA (frontend/) en web/dist/ para que Go lo embeba
	cd frontend && npm ci && npm run build
	@touch web/dist/.gitkeep   # vite --emptyOutDir lo borra; //go:embed lo necesita

.PHONY: build
build: frontend ## Compila el frontend y el servidor HTTP en bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(SERVER_BIN) ./cmd/server

.PHONY: dev
dev: ## Arranca el servidor Go (usa el último web/dist compilado)
	CONTROL_IPV_ENV=dev CONTROL_IPV_LOG_LEVEL=debug $(GO) run ./cmd/server

.PHONY: dev-front
dev-front: ## Arranca Vite con proxy /api -> :5175 (usar junto con `make dev`)
	cd frontend && npm run dev

# --- Escritorio (Wails). Requiere el CLI: go install github.com/wailsapp/wails/v2/cmd/wails@latest ---
# Compilar cmd/desktop arrastra WebKit/GTK (Linux) vía CGO. En macOS/Windows no
# hacen falta paquetes extra; el CI de Linux los instala.

.PHONY: wails-dev
wails-dev: ## Arranca la app de escritorio en modo desarrollo
	cd cmd/desktop && wails dev

.PHONY: wails-build
wails-build: ## Compila el ejecutable de escritorio (en cmd/desktop/build/bin)
	cd cmd/desktop && wails build -clean

.PHONY: wails-doctor
wails-doctor: ## Verifica el entorno de Wails
	wails doctor

.PHONY: desktop-compile
desktop-compile: ## Solo comprueba que cmd/desktop compila (sin empaquetar)
	$(GO) build -o /dev/null ./cmd/desktop

.PHONY: test
test: ## Ejecuta los tests
	$(GO) test ./...

.PHONY: vet
vet: ## go vet sobre todo el módulo
	$(GO) vet ./...

.PHONY: lint
lint: ## golangci-lint (si está instalado)
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run \
		|| echo "golangci-lint no instalado; omitiendo (opcional)"

.PHONY: check
check: tidy vet test ## tidy + vet + test

.PHONY: clean
clean: ## Borra artefactos de compilación
	rm -rf $(BIN_DIR)

# Los artefactos de paridad (migration/goldens/, migration/schema_actual.sql)
# están congelados: capturan el comportamiento de la versión Python 0.1.0.
# El generador vivía en backend/ (eliminado); recuperable del historial git
# si alguna vez hiciera falta regenerarlos.
