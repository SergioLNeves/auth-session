PRIVATE_KEY = private-key.pem
PUBLIC_KEY = public-key.pem
KEY_SIZE = 2048
DOCKER_IMAGE = auth-session

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  setup     - Install project dependencies and development tools"
	@echo "  run       - Run the application"
	@echo "  gen-key   - Generate RSA key pair for authentication"
	@echo "  mocks     - Generate mock implementations for testing"
	@echo "  lint      - Run code linter"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  help      - Show this help message"

.PHONY: setup
setup:
	go get ./...
	@go install github.com/vektra/mockery/v3@v3.6.3
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0
	@go install github.com/air-verse/air@latest
	make gen-key

.PHONY: run
run:
	@air

.PHONY: gen-key
gen-key:
	@echo "Gerando chave privada..."
	@openssl genrsa -out $(PRIVATE_KEY) $(KEY_SIZE)
	@chmod 600 $(PRIVATE_KEY)
	@echo "Extraindo chave pública..."
	@openssl rsa -in $(PRIVATE_KEY) -pubout -out $(PUBLIC_KEY)

.PHONY: mocks
mocks:
	@mockery

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: docker-build
docker-build:
	docker build -f docker/Dockerfile -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 -v $(PWD)/data:/data $(DOCKER_IMAGE)
