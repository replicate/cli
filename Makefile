SHELL := /bin/bash

DESTDIR ?=
PREFIX = /usr/local
BINDIR = $(PREFIX)/bin

INSTALL := install -m 0755
INSTALL_PROGRAM := $(INSTALL)

GO := go
GOOS := $(shell $(GO) env GOOS)
GOARCH := $(shell $(GO) env GOARCH)

VHS := vhs

default: all

.PHONY: all
all: replicate

replicate:
	CGO_ENABLED=0 $(GO) build -o $@ \
		-ldflags "-X github.com/replicate/cli/internal/cmd.version=$(REPLICATE_CLI_VERSION) -w" \
		main.go

demo.gif: replicate demo.tape
	PATH=$(PWD):$(PATH) $(VHS) demo.tape

.PHONY: install
install: replicate
	$(INSTALL_PROGRAM) -d $(DESTDIR)$(BINDIR)
	$(INSTALL_PROGRAM) replicate $(DESTDIR)$(BINDIR)/replicate

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(BINDIR)/replicate

.PHONY: clean
clean:
	$(GO) clean
	rm -f replicate

.PHONY: test
test:
	$(GO) get gotest.tools/gotestsum
	$(GO) run gotest.tools/gotestsum -- -timeout 1200s -parallel 5 ./...

.PHONY: lint
lint:
	$(GO) get github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

