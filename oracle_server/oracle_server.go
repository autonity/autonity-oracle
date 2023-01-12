package oracleserver

import (
	"autonity-oracle/aggregator"
	cryptoprovider "autonity-oracle/plugin_client"
	pricepool "autonity-oracle/price_pool"
	"autonity-oracle/types"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"io/fs"
	"io/ioutil" // nolint
	o "os"
	"sync"
	"time"
)

var (
	ValidDataAge            = 3 * 60 * 1000 // 3 minutes, data fetched within 3 minutes are valid to update the price.
	Version                 = "0.0.1"
	UpdateInterval          = 10 * time.Second // 10s, the data fetching interval for the oracle server's jobTicker job.
	PluginDiscoveryInterval = 2 * time.Second  // 2s, the plugin discovery interval.
)

type OracleServer struct {
	version string
	logger  hclog.Logger

	lock sync.RWMutex

	prices types.PriceBySymbol // aggregated prices which is referred by http data service to provide the data service.

	doneCh          chan struct{}
	jobTicker       *time.Ticker // the clock source to trigger the 10s interval job.
	discoveryTicker *time.Ticker

	pluginDIR         string                        // the dir saves the plugins.
	symbols           []string                      // the symbols for data fetching in oracle service.
	aggregator        types.Aggregator              // the price aggregator once we have multiple data providers.
	priceProviderPool *pricepool.PriceProviderPool  // the price pool organized by plugin_client and by symbols
	pluginClients     map[string]types.PluginClient // the plugin clients that connect with different adapters.
}

func NewOracleServer(symbols []string, pluginDir string) *OracleServer {
	os := &OracleServer{
		version:           Version,
		symbols:           symbols,
		pluginDIR:         pluginDir,
		prices:            make(types.PriceBySymbol),
		pluginClients:     make(map[string]types.PluginClient),
		doneCh:            make(chan struct{}),
		jobTicker:         time.NewTicker(UpdateInterval),
		discoveryTicker:   time.NewTicker(PluginDiscoveryInterval),
		aggregator:        aggregator.NewAveragePriceAggregator(),
		priceProviderPool: pricepool.NewPriceProviderPool(),
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String(),
		Output: o.Stdout,
		Level:  hclog.Debug,
	})

	// discover plugins from plugin dir at startup.
	binaries := os.listPluginDIR()
	if len(binaries) == 0 {
		// to stop the service on the start once there is no plugin in the db.
		panic(fmt.Sprintf("No plugins at plugin dir: %s, please build the plugins", os.pluginDIR))
	}
	for _, file := range binaries {
		pluginClient := cryptoprovider.NewPluginClient(file.Name(), pluginDir)
		pool := os.priceProviderPool.AddPriceProvider(file.Name())
		os.pluginClients[file.Name()] = pluginClient
		pluginClient.Initialize(pool)
	}

	return os
}

func (os *OracleServer) Version() string {
	return os.version
}

func (os *OracleServer) UpdateSymbols(symbols []string) {
	os.symbols = symbols
}

func (os *OracleServer) Symbols() []string {
	return os.symbols
}

func (os *OracleServer) GetPrices() types.PriceBySymbol {
	os.lock.RLock()
	defer os.lock.RUnlock()
	return os.prices
}

func (os *OracleServer) UpdatePrice(price types.Price) {
	os.lock.Lock()
	defer os.lock.Unlock()
	os.prices[price.Symbol] = price
}

func (os *OracleServer) UpdatePrices() {
	wg := &errgroup.Group{}
	for _, p := range os.pluginClients {
		plugin := p
		wg.Go(func() error {
			return plugin.FetchPrices(os.symbols)
		})
	}
	err := wg.Wait()
	if err != nil {
		os.logger.Error("fetching prices from plugins error: ", err.Error())
	}

	now := time.Now().UnixMilli()

	for _, s := range os.symbols {
		var prices []decimal.Decimal
		for _, plugin := range os.pluginClients {
			p, err := os.priceProviderPool.GetPriceProvider(plugin.Name()).GetPrice(s)
			if err != nil {
				continue
			}
			// only those price collected within 3 minutes are valid.
			if now-p.Timestamp < int64(ValidDataAge) && now >= p.Timestamp {
				prices = append(prices, p.Price)
			}
		}

		if len(prices) == 0 {
			continue
		}

		price := types.Price{
			Timestamp: now,
			Price:     prices[0],
			Symbol:    s,
		}

		if len(prices) > 1 {
			p, err := os.aggregator.Aggregate(prices)
			if err != nil {
				continue
			}
			price.Price = p
		}

		os.UpdatePrice(price)
	}
}

func (os *OracleServer) Stop() {
	os.doneCh <- struct{}{}
	for _, c := range os.pluginClients {
		p := c
		p.Close()
	}
}

func (os *OracleServer) Start() {
	// start the ticker job to fetch prices for all the symbols from all pluginClients on every 10s.
	// start the ticker jot to discover plugins on every 2s.
	for {
		select {
		case <-os.doneCh:
			os.discoveryTicker.Stop()
			os.jobTicker.Stop()
			os.logger.Info("the jobTicker jobs of oracle service is stopped")
			return
		case <-os.discoveryTicker.C:
			os.PluginRuntimeDiscovery()
		case <-os.jobTicker.C:
			os.UpdatePrices()
		}
	}
}

func (os *OracleServer) PluginRuntimeDiscovery() {
	binaries := os.listPluginDIR()
	os.logger.Debug("getting plugins from store: ", binaries)
	for _, file := range binaries {
		plugin, ok := os.pluginClients[file.Name()]
		if !ok {
			os.logger.Info("New plugin set up by oracle service: ", file)
			pluginClient := cryptoprovider.NewPluginClient(file.Name(), os.pluginDIR)
			pool := os.priceProviderPool.AddPriceProvider(file.Name())
			os.pluginClients[file.Name()] = pluginClient
			pluginClient.Initialize(pool)
			continue
		}
		// the plugin was created, now we check the modification time is after the creation time of the plugin,
		// and try to replace the old plugin if we have to.
		if file.ModTime().After(plugin.StartTime()) {
			os.logger.Info("Replacing plugin: ", file.Name())
			pricePool := os.priceProviderPool.GetPriceProvider(file.Name())
			// stop the former plugins process, and disconnect the net rpc connection.
			plugin.Close()
			// release the former client from oracle service.
			delete(os.pluginClients, file.Name())
			pluginClient := cryptoprovider.NewPluginClient(file.Name(), os.pluginDIR)
			os.pluginClients[file.Name()] = pluginClient
			pluginClient.Initialize(pricePool)
			os.logger.Info("Finnish the replacement of plugin: ", file.Name())
		}
	}
}

func (os *OracleServer) listPluginDIR() []fs.FileInfo {
	var plugins []fs.FileInfo

	files, err := ioutil.ReadDir(os.pluginDIR)
	if err != nil {
		os.logger.Error("cannot read from plugin store, please double check plugins are saved in the directory: ",
			os.pluginDIR, err.Error())
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		plugins = append(plugins, file)
	}
	return plugins
}
