# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: autoracle test e2e_test clean lint dep all

SOLC_VERSION = 0.8.2
BINDIR = ./build/bin
E2ETESTDIR = ./e2e_test
SOLC_BINARY = $(BINDIR)/solc_static_linux_v$(SOLC_VERSION)
PLUGINDIR = ./build/bin/plugins
SIMULATORBINDIR = ./data_source_simulator/build/bin
SIMULATORSRCDIR = ./data_source_simulator/binance_simulator
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
	go build -o $(E2ETESTDIR)/autoracle
	chmod +x $(BINDIR)/autoracle
	chmod +x $(E2ETESTDIR)/autoracle
	go build -o $(PLUGINDIR)/binance $(PLUGINSRCDIR)/binance/binance.go
	chmod +x $(PLUGINDIR)/*
	mkdir -p $(PLUGINSRCDIR)/fakeplugin/bin
	go build -o $(PLUGINSRCDIR)/fakeplugin/bin/fakeplugin $(PLUGINSRCDIR)/fakeplugin/fakeplugin.go
	go build -o $(E2ETESTDIR)/plugin_dir/fakeplugin $(PLUGINSRCDIR)/fakeplugin/fakeplugin.go
	go build -o $(E2ETESTDIR)/binance_plugin_dir/binance $(PLUGINSRCDIR)/binance/binance.go
	chmod +x $(PLUGINSRCDIR)/fakeplugin/bin/fakeplugin
	chmod +x $(E2ETESTDIR)/plugin_dir/fakeplugin
	@echo "Done building."
	@echo "Run \"$(BINDIR)/autoracle\" to launch autonity oracle."

oracle-contract:
	mkdir -p $(BINDIR)
	wget -O $(SOLC_BINARY) https://github.com/ethereum/solidity/releases/download/v$(SOLC_VERSION)/solc-static-linux
	chmod +x $(SOLC_BINARY)

simulator:
	mkdir -p $(SIMULATORBINDIR)
	go build -o $(SIMULATORBINDIR)/simulator $(SIMULATORSRCDIR)/main.go

clean:
	go clean -cache
	rm -rf build/_workspace/pkg $(BINDIR)/*

test: autoracle
	go test ./... -coverprofile=coverage.out

e2e-test: autoracle
	go test ./e2e_test/

dep:
	go mod download

lint:
	@./.github/tools/golangci-lint run --config ./.golangci.yml

mock:
	mockgen -package=mock -source=contract_binder/contract/interface.go > contract_binder/contract/mock/contract_mock.go
	mockgen -package=mock -source=types/interface.go > types/mock/l1_mock.go
all: autoracle lint test