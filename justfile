# minion-monitoring justfile

# Build the application
build:
    @mkdir -p bin
    go build -o bin/minion-mon .
    @echo "Build complete: bin/minion-mon"

# Run go vet
vet:
    go vet ./...

# Format the code
fmt:
    go fmt ./...

# Run the application (default to hosts.yaml)
run: build
    ./bin/minion-mon

# Run with verbose and markdown format
report: build
    ./bin/minion-mon --format markdown -v

# Clean build artifacts
clean:
    rm -rf bin/
    @echo "Cleaned bin/ folder"

# Install dependencies
deps:
    go mod tidy
    go get github.com/spf13/cobra
    go get gopkg.in/yaml.v3
    go get github.com/pelletier/go-toml/v2
    go get golang.org/x/crypto/ssh
