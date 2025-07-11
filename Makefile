# Makefile simple pour todo-cli-go
.PHONY: build test clean run help install

# Variables
BINARY_NAME=todo
MAIN_FILES=main.go import.go
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT = $(shell git rev-parse HEAD)

# Flags de build avec injection des variables
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Commandes principales
build: ## Compiler le binaire
	#go build -o $(BINARY_NAME) $(MAIN_FILES)
	go build $(LDFLAGS) -o todo main.go import.go

version: ## Afficher la version qui sera compilée
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

test: ## Lancer tous les tests
	go test -v

test-short: ## Tests rapides (sans stress tests)
	go test -v -short

test-cli: ## Tests CLI uniquement
	go test -v -run "TestCLI"

test-unit: ## Tests unitaires uniquement
	go test -v -run "TestTodoManager|TestGenerate|TestParse|TestValidate|TestFilter"

test-e2e: ## Tests end-to-end uniquement
	go test -v -run "TestE2E"

test-coverage: ## Tests avec couverture de code
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Rapport de couverture généré: coverage.html"

run: build ## Compiler et lancer l'application
	./$(BINARY_NAME)

install: build ## Installer le binaire dans $GOPATH/bin
	cp $(BINARY_NAME) $(GOPATH)/bin/

clean: ## Nettoyer les fichiers générés
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f *.csv

help: ## Afficher cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Commandes de développement
dev: ## Mode développement avec rebuild automatique
	@echo "Mode développement - Ctrl+C pour arrêter"
	@while inotifywait -e modify *.go 2>/dev/null; do make build && echo "✅ Rebuild terminé"; done

lint: ## Vérification du code (si golangci-lint installé)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go vet ./...; \
		go fmt ./...; \
	fi

bench: ## Benchmarks de performance
	go test -bench=. -benchmem

# Commandes pratiques
demo: build ## Démonstration rapide de l'application
	@echo "🚀 Démonstration de todo-cli-go:"
	@echo ""
	./$(BINARY_NAME) add "Tâche de démonstration" +demo --priority=high
	./$(BINARY_NAME) add "Autre tâche" @maison --priority=medium
	./$(BINARY_NAME) list
	@echo ""
	@echo "✅ Démonstration terminée!"

check: ## Vérification complète avant commit
	make lint
	make test-short
	@echo "✅ Vérifications terminées - prêt pour commit!"

# Par défaut, afficher l'aide
.DEFAULT_GOAL := help
