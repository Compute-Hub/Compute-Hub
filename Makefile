# Go parameters
GO        := go
BINARY    := Compute-Hub
MAIN      := cmd/compute-hub/main.go
OS        := $(shell go env GOOS)

# Default target
.PHONY: all
all: build_all

# Build the Go binary for all Operating Systems
.PHONY: build_all
build_all:
	@echo "Building $(BINARY) for all OS..."
	GOARCH=amd64 GOOS=linux $(GO) build -o ./build/$(BINARY)-linux $(MAIN)
	GOARCH=amd64 GOOS=windows $(GO) build -o ./build/$(BINARY)-windows $(MAIN)
	GOARCH=amd64 GOOS=darwin $(GO) build -o ./build/$(BINARY)-mac $(MAIN)

# Build the Go binary for Current Operating System
.PHONY: build
build:
	@echo "Building $(BINARY) for $(OS)..."
	GOARCH=amd64 GOOS=$(OS) $(GO) build -o ./build/$(BINARY)-$(OS) $(MAIN)

# Run the Go application
.PHONY: run
run: build
	@echo "Running $(BINARY) for $(OS)..."
	./build/$(BINARY)-$(OS)

# Clean the project
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GO) clean
	@rm -f $(BINARY)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build   - Build the Go binary"
	@echo "  make run     - Run the Go application"
	@echo "  make clean   - Clean the project"
