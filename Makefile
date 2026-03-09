# Gistsync Makefile

BINARY_NAME=gistsync
VERSION=$(shell cat VERSION 2>/dev/null || echo "dev")
GIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
INSTALL_DIR=/usr/local/bin

.PHONY: build install clean local-install help dev install-tools install-hooks test

help:
	@echo "Gistsync Build & Install Commands"
	@echo "--------------------------------"
	@echo "make build         - Local build for current OS/Arch"
	@echo "make install       - Install to /usr/local/bin (requires sudo on Unix)"
	@echo "make local-install - Install to ~/go/bin (via go install)"
	@echo "make dev           - Live reload development mode (rebuilds and installs on change)"
	@echo "make install-tools - Install development tools (e.g., air)"
	@echo "make install-hooks - Install Git hooks for versioning and changelog"
	@echo "make clean         - Remove binaries and dist/ directory"
	@echo "make test          - Run the automated bash test suite"

test:
	@chmod +x tests/*.sh
	./tests/run_tests.sh

build:
	go build -ldflags="-X 'github.com/karanshah229/gistsync/cmd.version=$(VERSION)-$(GIT_HASH)'" -o $(BINARY_NAME) main.go

install: build
ifeq ($(OS),Windows_NT)
	@echo "Windows detected. Installing via 'go install' to ensure %GOPATH%/bin is used..."
	go install .
else
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR) (macOS/Linux)..."
	@if [ -w $(INSTALL_DIR) ]; then \
		cp $(BINARY_NAME) $(INSTALL_DIR)/; \
	else \
		sudo cp $(BINARY_NAME) $(INSTALL_DIR)/; \
	fi
	@echo "Successfully installed $(BINARY_NAME)"
endif

local-install:
	go install .

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

dev: install-tools install-hooks
	@echo "Warming up sudo for live-reload session..."
	@sudo -v
	@# Background loop to keep sudo session alive
	@echo "Starting live-reload with sudo keep-alive..."
	@(while true; do sudo -n true; sleep 60; kill -0 $$$$ || exit; done 2>/dev/null &) && go tool air

install-tools:
	go mod tidy

install-hooks:
	@echo "Installing Git hooks..."
	@mkdir -p .git/hooks
	@cp scripts/git-hooks/pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Hooks installed successfully!"
