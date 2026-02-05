.PHONY: build dev clean

# Build the application
build:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Building backend..."
	go build -o build/bin/gridea-pro .

# Clean build artifacts
clean:
	rm -rf build/bin
	rm -rf backend/cmd/gridea-pro/frontend
	rm -rf backend/cmd/gridea-pro/main

# Run development mode (using Wails default, might need config tweaking for moved main.go)
dev:
	wails dev
