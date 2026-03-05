# Gistsync Makefile

BINARY_NAME=gistsync
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")
INSTALL_DIR=/usr/local/bin

.PHONY: build install clean local-install help

help:
	@echo "Gistsync Build & Install Commands"
	@echo "--------------------------------"
	@echo "make build         - Local build for current OS/Arch"
	@echo "make install       - Install to /usr/local/bin (requires sudo on Unix)"
	@echo "make local-install - Install to ~/go/bin (via go install)"
	@echo "make clean         - Remove binaries and dist/ directory"

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o $(BINARY_NAME) main.go

install: build
ifeq ($(OS),Windows_NT)
	@echo "Windows detected. Installing via 'go install' to ensure %GOPATH%/bin is used..."
	go install .
else
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR) (macOS/Linux)..."
	@if [ -w $(INSTALL_DIR) ]; then \
		mv $(BINARY_NAME) $(INSTALL_DIR)/; \
	else \
		sudo mv $(BINARY_NAME) $(INSTALL_DIR)/; \
	fi
	@echo "Successfully installed $(BINARY_NAME)"
endif

local-install:
	go install .

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
