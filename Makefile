# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: autoracle test e2e_test clean lint dep all

SOLC_VERSION = 0.8.2
BIN_DIR = ./build/bin
E2E_TEST_DIR = ./e2e_test
E2E_TEST_PLUGIN_DIR = $(E2E_TEST_DIR)/plugins
E2E_TEST_TEMPLATE_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/template_plugins
E2E_TEST_PRD_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/production_plugins
E2E_TEST_SML_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/simulator_plugins
E2E_TEST_MIX_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/mix_plugins
E2E_TEST_FOREX_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/forex_plugins
SOLC_BINARY = $(BIN_DIR)/solc_static_linux_v$(SOLC_VERSION)
PLUGIN_DIR = ./build/bin/plugins
SIMULATOR_BIN_DIR = ./data_source_simulator/build/bin
SIMULATOR_SRC_DIR = ./data_source_simulator/binance_simulator
PLUGIN_SRC_DIR = ./plugins
DOCKER_SUDO = $(shell [ `id -u` -eq 0 ] || id -nG $(USER) | grep "\<docker\>" > /dev/null || echo sudo )
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
	mkdir -p $(E2E_TEST_TEMPLATE_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_PRD_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_SML_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_MIX_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_FOREX_PLUGIN_DIR)

    # build oracle client
	go build -o $(BIN_DIR)/autoracle
	chmod +x $(BIN_DIR)/autoracle
	cp $(BIN_DIR)/autoracle $(E2E_TEST_DIR)/autoracle

    # build production plugins
	go build -o $(PLUGIN_DIR)/binance $(PLUGIN_SRC_DIR)/binance/binance.go
	go build -o $(PLUGIN_DIR)/simulator_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/simulator_plugin.go
	go build -o $(PLUGIN_DIR)/forex_currencyfreaks $(PLUGIN_SRC_DIR)/forex_currencyfreaks/forex_currencyfreaks.go
	go build -o $(PLUGIN_DIR)/forex_currencylayer $(PLUGIN_SRC_DIR)/forex_currencylayer/forex_currencylayer.go
	go build -o $(PLUGIN_DIR)/forex_exchangerate $(PLUGIN_SRC_DIR)/forex_exchangerate/forex_exchangerate.go
	go build -o $(PLUGIN_DIR)/forex_openexchange $(PLUGIN_SRC_DIR)/forex_openexchange/forex_openexchange.go
	chmod +x $(PLUGIN_DIR)/*

    # build template plugin for integration test
	mkdir -p $(PLUGIN_SRC_DIR)/template_plugin/bin
	go build -o $(PLUGIN_SRC_DIR)/template_plugin/bin/template_plugin $(PLUGIN_SRC_DIR)/template_plugin/template_plugin.go
	chmod +x $(PLUGIN_SRC_DIR)/template_plugin/bin/template_plugin

    # build simulator for integration test
	go build -o $(SIMULATOR_BIN_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go
	chmod +x $(SIMULATOR_BIN_DIR)/simulator
	cp $(SIMULATOR_BIN_DIR)/simulator $(E2E_TEST_DIR)/simulator

    # cp plugins for e2e testing
	cp $(PLUGIN_SRC_DIR)/template_plugin/bin/template_plugin $(E2E_TEST_MIX_PLUGIN_DIR)/template_plugin
	cp $(PLUGIN_SRC_DIR)/template_plugin/bin/template_plugin $(E2E_TEST_TEMPLATE_PLUGIN_DIR)/template_plugin

	cp $(PLUGIN_DIR)/binance $(E2E_TEST_PRD_PLUGIN_DIR)/binance

	# cp forex plugins for e2e testing
	cp $(PLUGIN_DIR)/forex_currencyfreaks $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_currencyfreaks
	cp $(PLUGIN_DIR)/forex_currencylayer $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_currencylayer
	cp $(PLUGIN_DIR)/forex_exchangerate $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_exchangerate
	cp $(PLUGIN_DIR)/forex_openexchange $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_openexchange

    # build simulator plugin
	go build -o $(E2E_TEST_SML_PLUGIN_DIR)/sim_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/simulator_plugin.go
	chmod +x $(E2E_TEST_SML_PLUGIN_DIR)/sim_plugin

	cp  $(E2E_TEST_SML_PLUGIN_DIR)/sim_plugin $(E2E_TEST_MIX_PLUGIN_DIR)/sim_plugin

	@echo "Done building."
	@echo "Run \"$(BIN_DIR)/autoracle\" to launch autonity oracle."

simulator:
	go build -o $(SIMULATOR_BIN_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go

oracle-contract:
	mkdir -p $(BIN_DIR)
	wget -O $(SOLC_BINARY) https://github.com/ethereum/solidity/releases/download/v$(SOLC_VERSION)/solc-static-linux
	chmod +x $(SOLC_BINARY)

# Builds the docker image and checks that we can run the autonity binary inside
# it.

build-docker-image:
	@$(DOCKER_SUDO) docker build -t autoracle .
	@$(DOCKER_SUDO) docker run --rm autoracle -h > /dev/null

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
