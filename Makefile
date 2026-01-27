BINARY_NAME=qs
VERSION=v0.1.0
INSTALL_PATH=/usr/local/bin

.PHONY: all build clean install uninstall release

all: build

build:
	@echo "Building QuotaSense CLI..."
	go build -o $(BINARY_NAME) main.go

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -rf dist/

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	@echo "Installation complete."

release:
	@echo "Building releases..."
	mkdir -p dist
	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME) main.go
	tar -czf dist/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz -C dist $(BINARY_NAME)
	# Darwin ARM64
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME) main.go
	tar -czf dist/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz -C dist $(BINARY_NAME)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME) main.go
	tar -czf dist/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz -C dist $(BINARY_NAME)
	@rm dist/$(BINARY_NAME)
	@echo "Releases are ready in dist/ folder."
