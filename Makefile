SHELL := /bin/bash
COMPOSE := docker compose
SERVICES := identity ingestion rules alert case audit metering reporting
TOOLS := migrate seed datasetgen
TF_DIRS := infra/terraform/org infra/terraform/bootstrap infra/terraform/envs/dev infra/terraform/envs/staging infra/terraform/envs/prod

.PHONY: help up web down ps logs build test test-race lint vet fmt tidy coverage migrate seed datasetgen fp-gate demo reset-demo clean ci tf-fmt tf-validate

help:
	@echo "up        start the full stack"
	@echo "web       start the optional Next.js web UI at http://localhost:3000"
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
	@echo "tf-fmt    format Terraform stacks"
	@echo "tf-validate validate Terraform stacks without cloud credentials"

up:
	$(COMPOSE) up --build -d
	@echo "Postgres 5432 | Identity 8081 | Ingestion 8082 | Rules 8083 | Alert 8084 | Case 8085 | Audit 8086 | Metering 8087 | Reporting 8088"
	@echo "Web UI optional: make web"

web:
	$(COMPOSE) --profile web up -d --build web
	@echo "Web UI http://localhost:3000 (login with a seeded demo user)"

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

tf-fmt:
	terraform fmt -recursive infra/terraform

tf-validate:
	@mkdir -p $(HOME)/.terraform.d/plugin-cache; \
	for dir in $(TF_DIRS); do \
		echo "validating $$dir"; \
		TF_PLUGIN_CACHE_DIR=$(HOME)/.terraform.d/plugin-cache terraform -chdir=$$dir init -backend=false -input=false >/dev/null || exit 1; \
		TF_PLUGIN_CACHE_DIR=$(HOME)/.terraform.d/plugin-cache terraform -chdir=$$dir validate || exit 1; \
	done

clean:
	rm -rf bin coverage.out coverage.html
