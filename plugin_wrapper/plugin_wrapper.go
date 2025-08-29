package pluginwrapper

import (
	"autonity-oracle/config"
	"autonity-oracle/types"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var (
	sampleTTL = 30 // 30s, the TTL of a sample before GC it.
)

// PluginWrapper is the unified wrapper for the interface of a plugin, it contains metadata of a corresponding
// plugin, buffers recent data samples measured from the corresponding plugin.
type PluginWrapper struct {
	version          string
	conf             *config.PluginConfig
	dataSrcType      types.DataSourceType
	lockService      sync.RWMutex
	lockSamples      sync.RWMutex
	samples          map[string]map[int64]types.Price
	latestTimestamps map[string]int64 // to track latest timestamps of samples

	plugin  *plugin.Client
	adapter types.Adapter
	name    string
	startAt time.Time
	logger  hclog.Logger

	doneCh         chan struct{}
	chSampleEvent  chan *types.SampleEvent
	subSampleEvent event.Subscription
	samplingSub    types.SampleEventSubscriber

	// metrics for the prices that are sampled by per plugin.
	priceMetrics map[string]metrics.GaugeFloat64
}

func NewPluginWrapper(logLevel hclog.Level, name string, pluginDir string, sub types.SampleEventSubscriber, conf *config.PluginConfig) *PluginWrapper {
	// Create a hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   name,
		Output: os.Stdout,
		Level:  logLevel,
	})

	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{},
	}

	// We're a host! Create the plugin life cycle object with configuration
	pg := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(fmt.Sprintf("%s/%s", pluginDir, name)), //nolint
		Logger:          logger,
	})

	p := &PluginWrapper{
		name:             name,
		plugin:           pg,
		conf:             conf,
		samplingSub:      sub,
		startAt:          time.Now(),
		doneCh:           make(chan struct{}),
		samples:          make(map[string]map[int64]types.Price),
		latestTimestamps: make(map[string]int64),
		chSampleEvent:    make(chan *types.SampleEvent),
		priceMetrics:     make(map[string]metrics.GaugeFloat64),
		logger:           logger,
	}

	return p
}

func (pw *PluginWrapper) Config() *config.PluginConfig {
	return pw.conf
}

func (pw *PluginWrapper) AddSample(prices []types.Price, ts int64) {
	pw.lockSamples.Lock()
	defer pw.lockSamples.Unlock()
	for _, p := range prices {
		tsMap, ok := pw.samples[p.Symbol]
		if !ok {
			tsMap = make(map[int64]types.Price)
			pw.samples[p.Symbol] = tsMap
		}
		tsMap[ts] = p

		// Update the latest timestamp
		if latestTs, exists := pw.latestTimestamps[p.Symbol]; !exists || ts > latestTs {
			pw.latestTimestamps[p.Symbol] = ts
		}
	}
}

// SelectSample returns the aggregated price computed from a set of pre-samples of a symbol by a specific plugin.
// For data points from AMM and AFQ markets, they are aggregated by the samples of the recent pre-samplings period,
// while for data points from CEX, the last sample of the pre-sampling period will be taken.
// The target is the timestamp on which the round block is mined, it's used to select datapoint from CEX data source.
func (pw *PluginWrapper) SelectSample(symbol string, target int64) (types.Price, error) {
	pw.lockSamples.RLock()
	defer pw.lockSamples.RUnlock()
	tsMap, ok := pw.samples[symbol]
	if !ok {
		return types.Price{}, types.ErrNoAvailablePrice
	}

	// Short-circuit if there's only one sample of data points from CEX
	if len(tsMap) == 1 {
		for _, price := range tsMap {
			return price, nil // Return the only sample
		}
	}

	// Try to get the target TS sample, otherwise we search for the nearest measurement.
	if p, ok := tsMap[target]; ok {
		return p, nil
	}

	var nearestKey int64
	var minDistance int64 = math.MaxInt64
	for ts := range tsMap {
		distance := target - ts
		if distance < 0 {
			distance = ts - target
		}

		if distance < minDistance {
			nearestKey = ts
			minDistance = distance
		}
	}

	price := tsMap[nearestKey]
	pw.logger.Debug("nearest sample", "symbol", symbol, "samples", len(tsMap), "targetTS", target, "nearestTS", nearestKey, "price", price)
	return price, nil
}

// GCExpiredSamples removes data points that are older than the TTL seconds of per plugin, it leaves recent samples
// together with next round's pre-samples as the input for the price aggregation for AMM, AFQ plugins. While, for CEX
// plugins, only the latest sample are kept without GC.
func (pw *PluginWrapper) GCExpiredSamples() {
	pw.lockSamples.Lock()
	defer pw.lockSamples.Unlock()

	currentTime := time.Now().Unix() // Get the current time in seconds
	threshold := currentTime - int64(sampleTTL)

	for symbol, tsMap := range pw.samples {
		if len(tsMap) == 0 {
			continue // Skip if there are no samples for this symbol
		}

		// Remove samples older than TTL seconds.
		for ts := range tsMap {
			if ts < threshold {
				delete(tsMap, ts)
			}
		}

		// If no samples left, remove the symbol from the map
		if len(tsMap) == 0 {
			delete(pw.samples, symbol)
			delete(pw.latestTimestamps, symbol) // Also clean up the latest timestamp
		}
	}
}

func (pw *PluginWrapper) Name() string {
	return pw.name
}

func (pw *PluginWrapper) Version() string {
	return pw.version
}

func (pw *PluginWrapper) StartTime() time.Time {
	return pw.startAt
}

// Initialize start the plugin, connect to it and do a handshake via state() interface.
func (pw *PluginWrapper) Initialize(chainID int64) error {
	// start the plugin process and connect to it
	rpcClient, err := pw.plugin.Client()
	if err != nil {
		pw.logger.Error("cannot start plugin process", "error", err.Error())
		return err
	}

	// dispenses a new instance of the plugin
	raw, err := rpcClient.Dispense("adapter")
	if err != nil {
		pw.logger.Error("cannot dispense adapter", "error", err.Error())
		return err
	}

	pw.adapter = raw.(types.Adapter)

	// load with plugin's statement, check if chainID is matched.
	state, err := pw.state(chainID)
	if err != nil {
		pw.logger.Error("cannot get plugin's pluginState", "error", err.Error())
		return err
	}
	pw.dataSrcType = state.DataSourceType
	pw.version = state.Version

	// create metrics for plugin on init phase.
	if metrics.Enabled {
		for _, symbol := range state.AvailableSymbols {
			name := strings.Join([]string{"oracle", pw.Name(), symbol, "price"}, "/")
			metrics.GetOrRegisterGaugeFloat64(name, nil)
		}
	}

	if state.KeyRequired && pw.conf.Key == "" {
		return types.ErrMissingServiceKey
	}

	// all good, start to subscribe data sampling event from oracle server, and listen for sampling.
	go pw.start()
	pw.logger.Info("plugin is up and running", "name", pw.name, "state", state)
	return nil
}

func (pw *PluginWrapper) Exited() bool {
	return pw.plugin.Exited()
}

func (pw *PluginWrapper) start() {
	pw.subSampleEvent = pw.samplingSub.WatchSampleEvent(pw.chSampleEvent)
	for {
		select {
		case <-pw.doneCh:
			pw.logger.Info("plugin is stopping...", "name", pw.name)
			return
		case err := <-pw.subSampleEvent.Err():
			if err != nil {
				pw.logger.Error("plugin wrapper main loop", "error", err.Error())
			}
			return
		case sampleEvent := <-pw.chSampleEvent:
			pw.logger.Debug("sampling price", "symbols", sampleEvent.Symbols, "TS", sampleEvent.TS)
			go func() {
				err := pw.fetchPrices(sampleEvent.Symbols, sampleEvent.TS)
				if err != nil {
					pw.logger.Warn("fetch price routine", "error", err.Error())
					return
				}
			}()
		}
	}
}

func (pw *PluginWrapper) state(chainID int64) (types.PluginStatement, error) {
	var s types.PluginStatement
	state, err := pw.adapter.State(chainID)
	if err != nil {
		return s, err
	}

	return state, nil
}

func (pw *PluginWrapper) fetchPrices(symbols []string, ts int64) error {
	// prevent race condition throughout data sampling routines in case of waiting for timeout.
	pw.lockService.Lock()
	defer pw.lockService.Unlock()

	report, err := pw.adapter.FetchPrices(symbols)
	if err != nil {
		return err
	}

	if len(report.UnRecognizableSymbols) != 0 {
		pw.logger.Debug("some symbol are not supported yet in this plugin", "unsupported", report.UnRecognizableSymbols)
	}

	if len(report.Prices) > 0 {
		pw.logger.Debug("sampled symbols", "data points", report.Prices)
		pw.AddSample(report.Prices, ts)
		if metrics.Enabled {
			pw.updateMetrics(report.Prices)
		}
	}
	return nil
}

func (pw *PluginWrapper) updateMetrics(prices []types.Price) {
	for _, p := range prices {
		m, ok := pw.priceMetrics[p.Symbol]
		if !ok {
			name := strings.Join([]string{"oracle", pw.Name(), p.Symbol, "price"}, "/")
			gauge := metrics.GetOrRegisterGaugeFloat64(name, nil)
			gauge.Update(p.Price.InexactFloat64())
			pw.priceMetrics[p.Symbol] = gauge
			continue
		}
		m.Update(p.Price.InexactFloat64())
	}
}

func (pw *PluginWrapper) CleanPluginProcess() {
	pw.plugin.Kill()
}

func (pw *PluginWrapper) Close() {
	pw.plugin.Kill()
	pw.doneCh <- struct{}{}
	pw.subSampleEvent.Unsubscribe()
	pw.logger.Info("plugin is stopped")
}
