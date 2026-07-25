.PHONY: fmt lint test build web compose-config check

fmt:
	gofmt -w agent server protocol

lint:
	cd agent && go vet ./...
	cd server && go vet ./...
	cd web && npm run lint

test:
	cd agent && go test ./...
	cd server && go test ./...
	cd web && npm test

build:
	cd agent && go build ./cmd/vpsentinel-agent
	cd server && go build ./cmd/vpsentinel-server
	cd web && npm run build

web:
	cd web && npm ci && npm run build

compose-config:
	docker compose --env-file .env.example config --quiet

check: lint test build compose-config
