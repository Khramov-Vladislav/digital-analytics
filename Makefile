# Makefile for digital-analytics resume generator

BINARY := app

.PHONY: all build run stop clean

# default target
all: build

# compile the Go binary
build:
	@echo "Building binary $(BINARY)..."
	go build -o $(BINARY) ./cmd
	@echo "Build OK"

# run the server (blocks)
run: build
	@echo "Starting server..."
	./$(BINARY)

# kill any running instance
stop:
	@echo "Stopping server if running..."
	pkill -f './$(BINARY)' 2>/dev/null || true

# remove compiled artifacts
clean: stop
	@echo "Cleaning up..."
	rm -f $(BINARY)
