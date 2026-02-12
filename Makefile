.PHONY: build test install clean
BINARY = awsctx
PREFIX ?= /usr/local
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -X github.com/yetmike/awsctx/internal/awsctx.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/awsctx

test:
	go test -v ./...

install: build
	install -m 0755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo ""
	@echo "Binary installed to $(PREFIX)/bin/$(BINARY)"
	@echo "For tab completions, add to your shell rc file:"
	@echo "  source $(CURDIR)/shell/awsctx.sh"

clean:
	rm -f $(BINARY)
