# Control IPV — migración a Go + Wails (camino A).
# Objetivos de desarrollo del backend Go. El frontend sigue en frontend/.

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

.PHONY: build
build: ## Compila el servidor HTTP en bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(SERVER_BIN) ./cmd/server

.PHONY: dev
dev: ## Arranca el servidor HTTP (CONTROL_IPV_ENV=dev)
	CONTROL_IPV_ENV=dev CONTROL_IPV_LOG_LEVEL=debug $(GO) run ./cmd/server

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

# --- Fase 0: artefactos de paridad (requieren backend/.venv) ---

VENV_PY := backend/.venv/bin/python

.PHONY: goldens
goldens: ## Regenera migration/goldens/*.json desde la versión Python
	$(VENV_PY) -m pytest backend/tests -q

.PHONY: schema-dump
schema-dump: ## Regenera migration/schema_actual.sql desde models.py
	$(VENV_PY) backend/tests/dump_schema.py

.PHONY: excel-fixtures
excel-fixtures: ## Regenera migration/goldens/fixtures/*.xlsx con el código Python
	$(VENV_PY) backend/tests/make_excel_fixtures.py
