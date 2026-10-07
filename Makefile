# INDICO Flash-Sale — one-liners for local dev and CI
#
# Usage:
#   make help          # list targets
#   make up            # start postgres + backend + frontend (Docker)
#   make db            # start only postgres
#   make backend       # run the Go backend natively (requires Postgres up)
#   make frontend      # run the Vite dev server
#   make test          # run unit + integration tests (race detector)
#   make stress        # run the no-overselling stress test
#   make down          # stop the stack
#   make clean         # stop + wipe volumes

# ---------- Config (override with `make up PORT=9090`) ----------
DATABASE_URL ?= postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable
PORT         ?= 8080
BACKEND_DIR  := backend
FRONTEND_DIR := frontend

# On Windows, allow `go test -race` by prepending MinGW to PATH if present.
ifdef OS
  MINGW_BIN := C:\Users\RIVAN\scoop\apps\mingw\current\bin
  ifeq ($(wildcard $(MINGW_BIN)),$(MINGW_BIN))
    PATH := $(MINGW_BIN);$(PATH)
  endif
endif

export DATABASE_URL
export PORT

.PHONY: help up down clean db backend frontend test test-unit test-integration test-docker stress stress-docker logs psql shell-backend shell-frontend build

help: ## show this help
	@echo INDICO Flash-Sale - Make targets
	@echo.
	@for /F "tokens=1,2,* delims=:#" %%a in ('findstr /B /R "^[a-zA-Z_-]*:.*##" $(MAKEFILE_LIST)') do @echo.   %%a  %%c

up: ## start postgres + backend + frontend (Docker)
	docker compose up -d --build
	@echo
	@echo "Frontend: http://localhost:5173"
	@echo "Backend:  http://localhost:8080/healthz"

db: ## start only Postgres
	docker compose up -d postgres
	@echo "Postgres ready on 127.0.0.1:5432"

down: ## stop the stack
	docker compose down

clean: ## stop and wipe Postgres data
	docker compose down -v

backend: db ## run the Go backend natively
	cd $(BACKEND_DIR) && go run ./cmd/server

frontend: ## run the Vite dev server
	cd $(FRONTEND_DIR) && npm install && npm run dev

build: ## build the Go binary
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -o bin/server ./cmd/server

test: test-unit test-integration ## run all tests (race detector)

test-unit: ## unit tests, no DB needed
	cd $(BACKEND_DIR) && go test -race -v -count=1 ./internal/service/...

test-integration: db ## integration + stress tests against real Postgres
	cd $(BACKEND_DIR) && go test -race -v -count=1 ./internal/repository/...

stress: db ## run only the no-overselling stress test (from host, requires CGO + reachable Postgres)
	cd $(BACKEND_DIR) && go test -race -v -count=1 -run TestStress_NoOverselling ./internal/repository/...

stress-docker: db ## run the stress test inside a one-shot container (no host CGO / no WSL networking hassles)
	docker run --rm --network indico_default -v "${CURDIR}/${BACKEND_DIR}:/src" -w /src golang:1.27-alpine \
		sh -c "apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 -run TestStress_NoOverselling ./internal/repository/..."

test-docker: db ## run the full test suite inside a one-shot container (CI-friendly)
	docker run --rm --network indico_default -v "${CURDIR}/${BACKEND_DIR}:/src" -w /src golang:1.27-alpine \
		sh -c "apk add --no-cache gcc musl-dev >/dev/null && DATABASE_URL='postgres://indico:indico@indico_postgres:5432/indico?sslmode=disable' go test -v -count=1 ./..."

logs: ## tail all container logs
	docker compose logs -f

psql: ## open psql in the postgres container
	docker exec -it indico_postgres psql -U indico -d indico

shell-backend: ## shell into the backend container
	docker exec -it indico_backend sh

shell-frontend: ## shell into the frontend container
	docker exec -it indico_frontend sh
