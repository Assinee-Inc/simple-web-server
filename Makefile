css:
	npx tailwindcss -i web/assets/css/input.css -o web/assets/css/tailwind.min.css --minify

css-watch:
	npx tailwindcss -i web/assets/css/input.css -o web/assets/css/tailwind.min.css --watch

run:
	go run cmd/web/main.go

test:
	go test ./...

test-e2e:
	npm run test:e2e

test-e2e-headed:
	npm run test:e2e:headed

install-e2e:
	npm install

setup-e2e: install-e2e
	npx cypress install

# Versioning commands
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT_HASH ?= $(shell git rev-parse HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build with version information
build:
	@echo "Building SimpleWebServer v$(VERSION)..."
	go build -ldflags "-X github.com/anglesson/simple-web-server/internal/config.Version=$(VERSION) -X github.com/anglesson/simple-web-server/internal/config.CommitHash=$(COMMIT_HASH) -X github.com/anglesson/simple-web-server/internal/config.BuildTime=$(BUILD_TIME)" -o bin/simple-web-server cmd/web/main.go

# Build for production
build-prod:
	@echo "Building SimpleWebServer for production..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/anglesson/simple-web-server/internal/config.Version=$(VERSION) -X github.com/anglesson/simple-web-server/internal/config.CommitHash=$(COMMIT_HASH) -X github.com/anglesson/simple-web-server/internal/config.BuildTime=$(BUILD_TIME)" -o bin/simple-web-server-linux-amd64 cmd/web/main.go

# Build for Heroku (with CGO enabled for PostgreSQL)
build-heroku:
	@echo "Building SimpleWebServer for Heroku..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/anglesson/simple-web-server/internal/config.Version=$(VERSION) -X github.com/anglesson/simple-web-server/internal/config.CommitHash=$(COMMIT_HASH) -X github.com/anglesson/simple-web-server/internal/config.BuildTime=$(BUILD_TIME)" -o bin/simple-web-server cmd/web/main.go

# Build for macOS
build-mac:
	@echo "Building SimpleWebServer for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X github.com/anglesson/simple-web-server/internal/config.Version=$(VERSION) -X github.com/anglesson/simple-web-server/internal/config.CommitHash=$(COMMIT_HASH) -X github.com/anglesson/simple-web-server/internal/config.BuildTime=$(BUILD_TIME)" -o bin/simple-web-server-darwin-amd64 cmd/web/main.go

# Build for Windows
build-windows:
	@echo "Building SimpleWebServer for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X github.com/anglesson/simple-web-server/internal/config.Version=$(VERSION) -X github.com/anglesson/simple-web-server/internal/config.CommitHash=$(COMMIT_HASH) -X github.com/anglesson/simple-web-server/internal/config.BuildTime=$(BUILD_TIME)" -o bin/simple-web-server-windows-amd64.exe cmd/web/main.go

# Build all platforms
build-all: build-prod build-mac build-windows

# Create a new version tag
tag:
	@echo "Creating tag v$(VERSION)..."
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

# Show current version info
version:
	@echo "Current version: $(VERSION)"
	@echo "Commit hash: $(COMMIT_HASH)"
	@echo "Build time: $(BUILD_TIME)"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run with hot reload (requires air)
dev:
	air

# Docker build
docker-build:
	docker build -t simple-web-server:$(VERSION) .

# Docker run
docker-run:
	docker run -p 8080:8080 simple-web-server:$(VERSION)

# Docker Compose commands for local development
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

dev: up
	air

# Security checks
security-check:
	@echo "🔒 Running security checks..."
	@echo "Checking for hardcoded credentials..."
	@grep -r "password.*=" internal/config/ || echo "✅ No hardcoded passwords found"
	@echo "Checking for insecure cookies..."
	@grep -r "Secure.*false" internal/ || echo "✅ No insecure cookies found"
	@grep -r "HttpOnly.*false" internal/ || echo "✅ No HttpOnly false cookies found"
	@echo "Checking for sensitive logs..."
	@grep -r "log.*token" internal/ || echo "✅ No token logging found"
	@echo "Checking for security headers..."
	@grep -r "X-Content-Type-Options" internal/ || echo "⚠️  Security headers not found"
	@echo "Checking for rate limiting..."
	@grep -r "RateLimit" internal/ || echo "⚠️  Rate limiting not found"
	@echo "✅ Security checks completed"

security-headers-test:
	@echo "🔒 Testing security headers..."
	@curl -s -I http://localhost:8080 | grep -E "(X-Content-Type-Options|X-Frame-Options|X-XSS-Protection)" || echo "⚠️  Security headers not found in response"

rate-limit-test:
	@echo "🔒 Testing rate limiting..."
	@echo "Making multiple requests to test rate limiting..."
	@for i in {1..10}; do \
		curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login; \
		echo " - Request $$i"; \
	done

# Environment setup
setup-env:
	@echo "🔧 Configurando arquivo .env..."
	@if [ -f ".env" ]; then \
		echo "⚠️  Arquivo .env já existe. Deseja sobrescrever? (y/N)"; \
		read -r response; \
		if [[ ! "$$response" =~ ^[Yy]$$ ]]; then \
			echo "❌ Operação cancelada."; \
			exit 1; \
		fi; \
	fi
	@cp env.template .env
	@echo "✅ Arquivo .env criado com sucesso!"
	@echo ""
	@echo "📝 Próximos passos:"
	@echo "1. Edite o arquivo .env com suas configurações"
	@echo "2. Configure as credenciais necessárias:"
	@echo "   - MAIL_USERNAME e MAIL_PASSWORD para email"
	@echo "   - S3_ACCESS_KEY e S3_SECRET_KEY para AWS S3"
	@echo "   - STRIPE_SECRET_KEY para pagamentos"
	@echo "   - HUB_DEVSENVOLVEDOR_TOKEN para validação de CPF"
	@echo ""
	@echo "🔒 IMPORTANTE: Nunca commite o arquivo .env no repositório!"
