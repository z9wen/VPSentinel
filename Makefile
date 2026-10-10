.PHONY: fmt lint test build web compose-config check terraform-check drill-up drill-down drill-console drill-password

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

TF_DRILL := terraform -chdir=deploy/terraform/drill

terraform-check:
	$(TF_DRILL) fmt -check -recursive
	$(TF_DRILL) init -backend=false -input=false
	$(TF_DRILL) validate
	$(TF_DRILL) test

# Disposable AWS drill environment; see docs/drills.md.
drill-up:
	$(TF_DRILL) init -input=false
	$(TF_DRILL) apply

drill-down:
	$(TF_DRILL) destroy

drill-console:
	@eval "$$($(TF_DRILL) output -raw console_tunnel_command)"

drill-password:
	@eval "$$($(TF_DRILL) output -raw admin_password_command)"
