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
}

func NewPluginWrapper(name string, pluginDir string, oracle types.SampleEventSubscriber) *PluginWrapper {
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   name,
		Output: os.Stdout,
		Level:  hclog.Debug,
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
		startAt:       time.Now(),
		doneCh:        make(chan struct{}),
		samples:       make(map[string]map[int64]types.Price),
		chSampleEvent: make(chan *types.SampleEvent),
		logger:        logger,
	}
	p.subSampleEvent = oracle.WatchSampleEvent(p.chSampleEvent)
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

	go pw.start()
	pw.logger.Info("plugin initialized", pw.name, state)
	return nil
}

func (pw *PluginWrapper) Exited() bool {
	return pw.plugin.Exited()
}

func (pw *PluginWrapper) start() {
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
				pw.logger.Debug("fetch price routine done successfully")
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
		pw.logger.Error("Fetch prices", "error", err.Error())
		return err
	}

	if len(report.BadSymbols) != 0 {
		pw.logger.Warn("find bad symbols: ", report.BadSymbols)
	}

	if len(report.Prices) > 0 {
		pw.AddSample(report.Prices, ts)
	}
	return nil
}

func (pw *PluginWrapper) Close() {
	pw.plugin.Kill()
	pw.doneCh <- struct{}{}
	pw.subSampleEvent.Unsubscribe()
}
