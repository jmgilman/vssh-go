gen:
	@echo "Running go generate..."
	go generate github.com/jmgilman/vssh/auth

test:
	@echo "Running all tests..."
	go test ./client/... ./auth/... ./internal/ui/...
