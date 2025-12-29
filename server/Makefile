.PHONY: help install dev build clean

help:
	@echo "Robo-Stream Server - Build Commands"
	@echo ""
	@echo "  make install    Install dependencies"
	@echo "  make dev        Run in development mode"
	@echo "  make build      Build for production"
	@echo "  make clean      Clean build artifacts"
	@echo ""

install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Downloading Go dependencies..."
	go mod download
	@echo "Done!"

dev:
	@echo "Starting development mode..."
	wails dev

build:
	@echo "Building for production..."
	wails build
	@echo "Build complete! Binary in build/bin/"

build-all:
	@echo "Building for all platforms..."
	wails build -platform darwin/arm64
	wails build -platform darwin/amd64
	wails build -platform windows/amd64
	wails build -platform linux/amd64
	@echo "All builds complete!"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
	@echo "Clean complete!"
