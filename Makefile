# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: autoracle test e2e_test clean lint dep all

SOLC_VERSION = 0.8.2
BIN_DIR = ./build/bin
E2E_TEST_DIR = ./e2e_test
E2E_TEST_PLUGIN_DIR = $(E2E_TEST_DIR)/plugins
E2E_TEST_FAKE_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/fake_plugins
E2E_TEST_PRD_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/production_plugins
E2E_TEST_SML_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/simulator_plugins
E2E_TEST_MIX_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/mix_plugins
SOLC_BINARY = $(BIN_DIR)/solc_static_linux_v$(SOLC_VERSION)
PLUGIN_DIR = ./build/bin/plugins
SIMULATOR_BIN_DIR = ./data_source_simulator/build/bin
SIMULATOR_SRC_DIR = ./data_source_simulator/binance_simulator
PLUGIN_SRC_DIR = ./plugins
GO ?= latest
LATEST_COMMIT ?= $(shell git log -n 1 master --pretty=format:"%H")
ifeq ($(LATEST_COMMIT),)
LATEST_COMMIT := $(shell git log -n 1 HEAD~1 --pretty=format:"%H")
endif

autoracle:
	mkdir -p $(BIN_DIR)
	mkdir -p $(PLUGIN_DIR)
	mkdir -p $(SIMULATOR_BIN_DIR)
	mkdir -p $(E2E_TEST_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_FAKE_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_PRD_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_SML_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_MIX_PLUGIN_DIR)
	go build -o $(BIN_DIR)/autoracle
	go build -o $(E2E_TEST_DIR)/autoracle
	chmod +x $(BIN_DIR)/autoracle
	chmod +x $(E2E_TEST_DIR)/autoracle

	go build -o $(PLUGIN_DIR)/binance $(PLUGIN_SRC_DIR)/binance/binance.go
	chmod +x $(PLUGIN_DIR)/*

	mkdir -p $(PLUGIN_SRC_DIR)/fakeplugin/bin
	go build -o $(PLUGIN_SRC_DIR)/fakeplugin/bin/fakeplugin $(PLUGIN_SRC_DIR)/fakeplugin/fakeplugin.go
	chmod +x $(PLUGIN_SRC_DIR)/fakeplugin/bin/fakeplugin

	go build -o $(SIMULATOR_BIN_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go
	go build -o $(E2E_TEST_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go
	chmod +x $(E2E_TEST_DIR)/simulator
	chmod +x $(SIMULATOR_BIN_DIR)/simulator

	go build -o $(E2E_TEST_MIX_PLUGIN_DIR)/fakeplugin $(PLUGIN_SRC_DIR)/fakeplugin/fakeplugin.go
	go build -o $(E2E_TEST_FAKE_PLUGIN_DIR)/fakeplugin $(PLUGIN_SRC_DIR)/fakeplugin/fakeplugin.go
	go build -o $(E2E_TEST_PRD_PLUGIN_DIR)/binance $(PLUGIN_SRC_DIR)/binance/binance.go
	go build -o $(E2E_TEST_SML_PLUGIN_DIR)/sim_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/simulator_plugin.go
	go build -o $(E2E_TEST_MIX_PLUGIN_DIR)/sim_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/simulator_plugin.go
	chmod +x $(E2E_TEST_MIX_PLUGIN_DIR)/fakeplugin
	chmod +x $(E2E_TEST_FAKE_PLUGIN_DIR)/fakeplugin
	chmod +x $(E2E_TEST_PRD_PLUGIN_DIR)/binance
	chmod +x $(E2E_TEST_SML_PLUGIN_DIR)/sim_plugin
	chmod +x $(E2E_TEST_MIX_PLUGIN_DIR)/sim_plugin

	@echo "Done building."
	@echo "Run \"$(BIN_DIR)/autoracle\" to launch autonity oracle."

simulator:
	go build -o $(SIMULATOR_BIN_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go

oracle-contract:
	mkdir -p $(BIN_DIR)
	wget -O $(SOLC_BINARY) https://github.com/ethereum/solidity/releases/download/v$(SOLC_VERSION)/solc-static-linux
	chmod +x $(SOLC_BINARY)

clean:
	go clean -cache
	rm -rf build/_workspace/pkg $(BIN_DIR)/*

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