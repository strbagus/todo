.PHONY: dev build test clean

# Run local development server (uses air if installed, otherwise go run)
dev:
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		go run ./cmd/server; \
	fi

# Build single binary
build:
	go build -o bin/todo ./cmd/server

# Run all tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build docker image
docker-build:
	docker build -t todo:latest .

# Deploy stack to Swarm
stack-deploy:
	docker stack deploy -c docker-compose.yml todo

