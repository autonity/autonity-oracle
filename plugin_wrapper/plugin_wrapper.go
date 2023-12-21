package pluginwrapper

import (
	"autonity-oracle/types"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"
	"sync"
	"time"
)

// PluginWrapper is the unified wrapper for the interface of a plugin, it contains metadata of a corresponding
// plugin, buffers recent data samples measured from the corresponding plugin.
type PluginWrapper struct {
	version     string
	conf        *types.PluginConfig
	lockService sync.RWMutex
	lockSamples sync.RWMutex
	samples     map[string]map[int64]types.Price
	plugin      *plugin.Client
	adapter     types.Adapter
	name        string
	startAt     time.Time
	logger      hclog.Logger

	doneCh         chan struct{}
	chSampleEvent  chan *types.SampleEvent
	subSampleEvent event.Subscription
	samplingSub    types.SampleEventSubscriber
}

func NewPluginWrapper(logLevel hclog.Level, name string, pluginDir string, sub types.SampleEventSubscriber, conf *types.PluginConfig) *PluginWrapper {
	// Create an hclog.Logger
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
		name:          name,
		plugin:        pg,
		conf:          conf,
		samplingSub:   sub,
		startAt:       time.Now(),
		doneCh:        make(chan struct{}),
		samples:       make(map[string]map[int64]types.Price),
		chSampleEvent: make(chan *types.SampleEvent),
		logger:        logger,
	}

	return p
}

func (pw *PluginWrapper) AddSample(prices []types.Price, ts int64) {
	pw.lockSamples.Lock()
	defer pw.lockSamples.Unlock()
	for _, p := range prices {
		tsMap, ok := pw.samples[p.Symbol]
		if !ok {
			tsMap = make(map[int64]types.Price)
			tsMap[ts] = p
			pw.samples[p.Symbol] = tsMap
			return
		}
		tsMap[ts] = p
	}
}

func (pw *PluginWrapper) GetSample(symbol string, target int64) (types.Price, error) {
	pw.lockSamples.RLock()
	defer pw.lockSamples.RUnlock()
	tsMap, ok := pw.samples[symbol]
	if !ok {
		return types.Price{}, types.ErrNoAvailablePrice
	}

	if p, ok := tsMap[target]; ok {
		return p, nil
	}

	// find return the nearest sampled price to the timestamp.
	var nearestKey int64
	var minDistance int64
	minDistance = math.MaxInt64
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

	return tsMap[nearestKey], nil
}

func (pw *PluginWrapper) GCSamples() {
	pw.lockSamples.Lock()
	defer pw.lockSamples.Unlock()
	for k := range pw.samples {
		delete(pw.samples, k)
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
func (pw *PluginWrapper) Initialize() error {
	// start the plugin process and connect to it
	rpcClient, err := pw.plugin.Client()
	if err != nil {
		pw.logger.Error("cannot start plugin process", err.Error())
		return err
	}

	// dispenses a new instance of the plugin
	raw, err := rpcClient.Dispense("adapter")
	if err != nil {
		pw.logger.Error("cannot dispense adapter", err.Error())
		return err
	}

	pw.adapter = raw.(types.Adapter)

	// load plugin's pluginState.
	state, err := pw.state()
	if err != nil {
		pw.logger.Error("cannot get plugin's pluginState")
		return err
	}
	pw.version = state.Version
	if state.KeyRequired && pw.conf.Key == "" {
		return types.ErrMissingServiceKey
	}

	// all good, start to subscribe data sampling event from oracle server, and listen for sampling.
	go pw.start()
	pw.logger.Info("plugin is up and running", pw.name, state)
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
			pw.logger.Info("plugin exist", "name", pw.name)
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
					pw.logger.Error("fetch price routine", "error", err.Error())
					return
				}
			}()
		}
	}
}

func (pw *PluginWrapper) state() (types.PluginState, error) {
	var s types.PluginState
	state, err := pw.adapter.State()
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
		pw.logger.Debug("the data source cannot recognize some symbol", "report", report)
	}

	if len(report.Prices) > 0 {
		pw.AddSample(report.Prices, ts)
	}
	return nil
}

func (pw *PluginWrapper) CleanPluginProcess() {
	pw.plugin.Kill()
}

func (pw *PluginWrapper) Close() {
	pw.plugin.Kill()
	pw.doneCh <- struct{}{}
	pw.subSampleEvent.Unsubscribe()
}
