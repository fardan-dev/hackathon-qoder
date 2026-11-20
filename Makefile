SHELL := /bin/bash
COMPOSE := docker compose
SERVICES := identity ingestion rules alert case audit metering reporting
TOOLS := migrate seed datasetgen

.PHONY: help up down ps logs build test test-race lint vet fmt tidy coverage migrate seed datasetgen fp-gate demo reset-demo clean ci

help:
	@echo "up        start the full stack"
	@echo "down      stop the stack (keep volumes)"
	@echo "reset     stop the stack and delete volumes"
	@echo "seed      seed demo users through the identity API"
	@echo "datasetgen generate deterministic synthetic datasets"
	@echo "fp-gate   run control-claim false-positive gate (writes backtest artifact)"
	@echo "demo      run the five-minute synthetic demo"
	@echo "reset-demo clear demo data but keep users and the rule catalog"
	@echo "migrate   apply database migrations locally"
	@echo "test      run unit tests"
	@echo "ci        vet + test + build all services"

up:
	$(COMPOSE) up --build -d
	@echo "Postgres 5432 | Identity 8081 | Ingestion 8082 | Rules 8083 | Alert 8084 | Case 8085 | Audit 8086 | Metering 8087 | Reporting 8088"

down:
	$(COMPOSE) down

reset:
	$(COMPOSE) down -v

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs -f --tail=200

build:
	@for svc in $(SERVICES); do \
		echo "building $$svc"; \
		go build -trimpath -o bin/$$svc ./services/$$svc/cmd/server || exit 1; \
	done

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint: vet
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; else echo "staticcheck not installed; go vet used"; fi

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

tidy:
	go mod tidy

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

migrate:
	$(COMPOSE) run --rm migrate

seed:
	go run ./cmd/seed

datasetgen:
	go run ./cmd/datasetgen -out datasets/synthetic

fp-gate:
	FP_GATE_OUT=backtest/fp-gate-report.json go test ./cmd/datasetgen -run 'TestControlClaimFalsePositiveGate|TestEngineOutputIndependentOfGroundTruth' -v

demo:
	./scripts/demo.sh

reset-demo:
	./scripts/reset-demo.sh

ci: vet test-race build

clean:
	rm -rf bin coverage.out coverage.html
