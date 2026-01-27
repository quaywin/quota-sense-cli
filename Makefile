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
	@echo "Launching release script..."
	./release.sh
