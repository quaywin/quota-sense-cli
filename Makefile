BINARY_NAME=qs
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
INSTALL_PATH=/usr/local/bin
LDFLAGS=-ldflags "-X github.com/quaywin/quota-sense-cli/cmd.Version=$(VERSION)"

.PHONY: all build clean install uninstall release

all: build

build:
	@echo "Building QuotaSense CLI $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -rf dist/

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	@echo "Installation complete."

release:
	@echo "Launching release script..."
	./release.sh
