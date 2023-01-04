# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: autoracle test e2e_test clean lint dep all

BINDIR = ./build/bin
GO ?= latest
LATEST_COMMIT ?= $(shell git log -n 1 master --pretty=format:"%H")
ifeq ($(LATEST_COMMIT),)
LATEST_COMMIT := $(shell git log -n 1 HEAD~1 --pretty=format:"%H")
endif

autoracle:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/autoracle
	chmod +x $(BINDIR)/autoracle
	@echo "Done building."
	@echo "Run \"$(BINDIR)/autoracle\" to launch autonity oracle."

clean:
	go clean -cache
	rm -rf build/_workspace/pkg $(BINDIR)/*

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

e2e_test:
	go test e2e_test/e2e_test.go

dep:
	go mod download

lint:
	@./.github/tools/golangci-lint run --config ./.golangci.yml

all: autoracle test test_coverage e2e_test lint



