BUF_VERSION := "1.30.1"
SQLC_VERSION := "1.26.0"

DOCKER_REGISTRY_PREFIX="christiancavuti"

include .env.dev
export $(shell sed 's/=.*//' .env.dev)

.PHONY: deps

sqlc:
	@echo "Downloading sqlc version ${SQLC_VERSION}"
	@mkdir -p ./bin
	@wget -qO- https://downloads.sqlc.dev/sqlc_${SQLC_VERSION}_linux_amd64.tar.gz | tar xz -C ./bin
	@chmod +x ./bin/sqlc

buf:
	@echo "Downloading buf version ${BUF_VERSION}"
	@mkdir -p ./bin
	@curl -s -L "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$$(uname -s)-$$(uname -m)" -o ./bin/buf
	@chmod +x ./bin/buf
	@./bin/buf mod update
	@mkdir -p ./proto

sqlc/directories:
	@echo "Creating sqlc directories"
	@mkdir -p db/migrations
	@mkdir -p db/queries

deps: sqlc sqlc/directories buf
	@go get github.com/pressly/goose/v3
	@go get go.uber.org/zap
	@go get github.com/jackc/pgx/v5
	@go get github.com/grpc-ecosystem/grpc-gateway/v2@2.19.1

generate:
	@./bin/sqlc generate
	@./bin/buf generate

init:
	@go mod init "github.com/Kavuti/$$(basename $$(pwd))"
	@$(MAKE) deps
	@mkdir -p ./service

docker/build:
	@docker build -t "${DOCKER_REGISTRY_PREFIX}/$$(basename $$(pwd))" .

docker/push:
	@docker push "${DOCKER_REGISTRY_PREFIX}/$$(basename $$(pwd))"


build:
	@go build -o ./cmd/$$(basename $$(pwd)) .

run/dev:
	@go run .