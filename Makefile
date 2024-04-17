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
		-ldflags "-X github.com/replicate/cli/internal.version=$(REPLICATE_CLI_VERSION) -w" \
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
	$(GO) test -v ./...

.PHONY: format
format:
	$(GO) run golang.org/x/tools/cmd/goimports@latest -d -w -local $(shell $(GO) list -m) .

.PHONY: lint
lint: lint-golangci lint-nilaway

.PHONY: lint-golangci
lint-golangci:
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2 run ./...

.PHONY: lint-nilaway
lint-nilaway:
	$(GO) run go.uber.org/nilaway/cmd/nilaway@v0.0.0-20240403175823-755a685ab68b ./...
