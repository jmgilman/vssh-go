gen:
	@echo "Running go generate..."
	go generate github.com/jmgilman/vssh/pkg/client
	go generate github.com/jmgilman/vssh/pkg/auth

test:
	@echo "Running all tests..."
	go test ./pkg/client/... ./pkg/auth/... ./internal/ui/...