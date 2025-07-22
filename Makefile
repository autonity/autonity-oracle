# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: mkdir oracle-server conf-file e2e-test-stuffs forex-plugins cex-plugins autoracle test e2e_test clean lint dep all

LINTER = ./bin/golangci-lint
GOLANGCI_LINT_VERSION = v1.62.0 # Change this to the desired version
SOLC_VERSION = 0.8.2
BIN_DIR = ./build/bin
CONF_FILE = ./config/oracle_config.yml
E2E_TEST_DIR = ./e2e_test
E2E_TEST_PLUGIN_DIR = $(E2E_TEST_DIR)/plugins
E2E_TEST_TEMPLATE_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/template_plugins
E2E_TEST_SML_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/simulator_plugins
E2E_TEST_OUTLIER_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/outlier_plugins
E2E_TEST_MIX_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/mix_plugins
E2E_TEST_FOREX_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/forex_plugins
E2E_TEST_CRYPTO_PLUGIN_DIR = $(E2E_TEST_PLUGIN_DIR)/crypto_plugins
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

# Download golangci-lint if not installed
.PHONY: install-linter
install-linter:
	@if [ ! -f $(LINTER) ]; then \
		echo "Downloading golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin $(GOLANGCI_LINT_VERSION); \
	fi

mkdir:
	mkdir -p $(BIN_DIR)
	mkdir -p $(PLUGIN_DIR)
	mkdir -p $(SIMULATOR_BIN_DIR)
	mkdir -p $(E2E_TEST_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_TEMPLATE_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_SML_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_MIX_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_FOREX_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_CRYPTO_PLUGIN_DIR)
	mkdir -p $(E2E_TEST_OUTLIER_PLUGIN_DIR)

oracle-server:
    # build oracle client
	go build -o $(BIN_DIR)/autoracle
	chmod +x $(BIN_DIR)/autoracle
	cp $(BIN_DIR)/autoracle $(E2E_TEST_DIR)/autoracle

conf-file:
	# copy example oracle_config.yml
	cp $(CONF_FILE) $(BIN_DIR)

e2e-test-stuffs:
    # build template plugin for e2e test and unit test.
	go build -o $(PLUGIN_SRC_DIR)/template_plugin/bin/template_plugin $(PLUGIN_SRC_DIR)/template_plugin/template_plugin.go
	go build -o $(E2E_TEST_MIX_PLUGIN_DIR)/template_plugin $(PLUGIN_SRC_DIR)/template_plugin/template_plugin.go
	go build -o $(E2E_TEST_TEMPLATE_PLUGIN_DIR)/template_plugin $(PLUGIN_SRC_DIR)/template_plugin/template_plugin.go
	chmod +x $(PLUGIN_SRC_DIR)/template_plugin/bin/template_plugin
	chmod +x $(E2E_TEST_MIX_PLUGIN_DIR)/template_plugin
	chmod +x $(E2E_TEST_TEMPLATE_PLUGIN_DIR)/template_plugin

    # build simulator for e2e test
	go build -o $(E2E_TEST_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go
	chmod +x $(E2E_TEST_DIR)/simulator

	# build amm plugin for e2e test.
	go build -o $(E2E_TEST_CRYPTO_PLUGIN_DIR)/crypto_uniswap $(PLUGIN_SRC_DIR)/crypto_uniswap/uniswap_usdcx/mainnet/crypto_uniswap_usdcx.go
	chmod +x $(E2E_TEST_CRYPTO_PLUGIN_DIR)/*

    # build mainnet simulator plugin for e2e test.
	go build -o $(E2E_TEST_SML_PLUGIN_DIR)/simulator_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/mainnet/simulator_plugin.go
	chmod +x $(E2E_TEST_SML_PLUGIN_DIR)/simulator_plugin
	cp  $(E2E_TEST_SML_PLUGIN_DIR)/simulator_plugin $(E2E_TEST_MIX_PLUGIN_DIR)/simulator_plugin

    # build outlier tester plugin for e2e test
	go build -o $(E2E_TEST_OUTLIER_PLUGIN_DIR)/outlier_plugin $(PLUGIN_SRC_DIR)/outlier_tester/outlier_tester.go
	chmod +x $(E2E_TEST_OUTLIER_PLUGIN_DIR)/outlier_plugin

    # cp forex plugins for e2e testing
	cp $(PLUGIN_DIR)/forex_currencyfreaks $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_currencyfreaks
	cp $(PLUGIN_DIR)/forex_currencylayer $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_currencylayer
	cp $(PLUGIN_DIR)/forex_exchangerate $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_exchangerate
	cp $(PLUGIN_DIR)/forex_openexchange $(E2E_TEST_FOREX_PLUGIN_DIR)/forex_openexchange

    # cp cex plugins for e2e testing
	cp $(PLUGIN_DIR)/crypto_coinbase $(E2E_TEST_CRYPTO_PLUGIN_DIR)/crypto_coinbase
	cp $(PLUGIN_DIR)/crypto_coingecko $(E2E_TEST_CRYPTO_PLUGIN_DIR)/crypto_coingecko
	cp $(PLUGIN_DIR)/crypto_kraken $(E2E_TEST_CRYPTO_PLUGIN_DIR)/crypto_kraken

# build ATN-USDC, NTN-USDC, NTN-ATN data point simulator binary
crypto_source_simulator:
	go build -o $(SIMULATOR_BIN_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go
	go build -o $(BIN_DIR)/simulator $(SIMULATOR_SRC_DIR)/main.go

forex-plugins:
	go build -o $(PLUGIN_DIR)/forex_currencyfreaks $(PLUGIN_SRC_DIR)/forex_currencyfreaks/forex_currencyfreaks.go
	go build -o $(PLUGIN_DIR)/forex_currencylayer $(PLUGIN_SRC_DIR)/forex_currencylayer/forex_currencylayer.go
	go build -o $(PLUGIN_DIR)/forex_exchangerate $(PLUGIN_SRC_DIR)/forex_exchangerate/forex_exchangerate.go
	go build -o $(PLUGIN_DIR)/forex_openexchange $(PLUGIN_SRC_DIR)/forex_openexchange/forex_openexchange.go
	go build -o $(PLUGIN_DIR)/forex_wise $(PLUGIN_SRC_DIR)/forex_wise/forex_wise.go	
	chmod +x $(PLUGIN_DIR)/*

cex-plugins:
	go build -o $(PLUGIN_DIR)/crypto_coinbase $(PLUGIN_SRC_DIR)/crypto_coinbase/crypto_coinbase.go
	go build -o $(PLUGIN_DIR)/crypto_coingecko $(PLUGIN_SRC_DIR)/crypto_coingecko/crypto_coingecko.go
	go build -o $(PLUGIN_DIR)/crypto_kraken $(PLUGIN_SRC_DIR)/crypto_kraken/crypto_kraken.go
	chmod +x $(PLUGIN_DIR)/*

# build amm plugins for bakerloo network:
amm-plugins-bakerloo:
	go build -o $(PLUGIN_DIR)/crypto_uniswap $(PLUGIN_SRC_DIR)/crypto_uniswap/uniswap_usdcx/bakerloo/crypto_uniswap_usdcx.go
	chmod +x $(PLUGIN_DIR)/*

# build amm plugins for main network:
amm-plugins-mainnet:
	go build -o $(PLUGIN_DIR)/crypto_uniswap $(PLUGIN_SRC_DIR)/crypto_uniswap/uniswap_usdcx/mainnet/crypto_uniswap_usdcx.go
	chmod +x $(PLUGIN_DIR)/*

# build simulator plugin for main network.
sim-plugin-mainnet:
	go build -o $(PLUGIN_DIR)/simulator_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/mainnet/simulator_plugin.go
	chmod +x $(PLUGIN_DIR)/simulator_plugin

# build simulator plugin for bakerloo network.
sim-plugin-bakerloo:
	go build -o $(PLUGIN_DIR)/simulator_plugin $(PLUGIN_SRC_DIR)/simulator_plugin/bakerloo/simulator_plugin.go
	chmod +x $(PLUGIN_DIR)/simulator_plugin

# build the whole components for autonity main network.
autoracle: mkdir oracle-server forex-plugins cex-plugins amm-plugins-mainnet conf-file e2e-test-stuffs
	@echo "Done oracle server and plugins building for autonity main network."
	@echo "Run \"$(BIN_DIR)/autoracle\" to launch autonity oracle server for autonity main network."

# build the whole components for bakerloo network.
autoracle-bakerloo: mkdir oracle-server forex-plugins cex-plugins amm-plugins-bakerloo conf-file
	@echo "Done oracle server and plugins building for autonity bakerloo network."
	@echo "Run \"$(BIN_DIR)/autoracle\" to launch autonity oracle server for autonity bakerloo network."

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

# Run the linter
lint: install-linter
	@$(LINTER) run --config ./.golangci.yml

mock:
	mockgen -package=mock -source=contract_binder/contract/interface.go > contract_binder/contract/mock/contract_mock.go
	mockgen -package=mock -source=types/interface.go > types/mock/l1_mock.go
all: autoracle lint test
