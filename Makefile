# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: autoracle test e2e_test clean lint dep all

BINDIR = ./build/bin
PLUGINDIR = ./build/bin/plugins
PLUGINSRCDIR = ./plugins
GO ?= latest
LATEST_COMMIT ?= $(shell git log -n 1 master --pretty=format:"%H")
ifeq ($(LATEST_COMMIT),)
LATEST_COMMIT := $(shell git log -n 1 HEAD~1 --pretty=format:"%H")
endif

autoracle:
	mkdir -p $(BINDIR)
	mkdir -p $(PLUGINDIR)
	go build -o $(BINDIR)/autoracle
	chmod +x $(BINDIR)/autoracle
	go build -o $(PLUGINDIR)/binance $(PLUGINSRCDIR)/binance/binance.go
	go build -o $(PLUGINDIR)/fakeplugin $(PLUGINSRCDIR)/fakeplugin/fakeplugin.go
	chmod +x $(PLUGINDIR)/*
	@echo "Done building."
	@echo "Run \"$(BINDIR)/autoracle\" to launch autonity oracle."

clean:
	go clean -cache
	rm -rf build/_workspace/pkg $(BINDIR)/*

test: autoracle
	go test ./...

test_coverage: autoracle
	go test ./... -coverprofile=coverage.out

e2e_test: autoracle
	go test e2e_test/e2e_test.go

dep:
	go mod download

lint:
	@./.github/tools/golangci-lint run --config ./.golangci.yml

all: autoracle test test_coverage e2e_test lint



