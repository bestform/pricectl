BINARY := pricectl
INSTALL_DIR := /usr/local/bin

.PHONY: build test install serve

build:
	go build -o $(BINARY) .

test:
	go test ./...

install: build
	mv $(BINARY) $(INSTALL_DIR)/$(BINARY)

serve: build
	./$(BINARY) serve
