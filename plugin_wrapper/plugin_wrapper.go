package pluginwrapper

import (
	"autonity-oracle/types"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"
	"sync"
	"time"
)

var errConnectionNotEstablished = errors.New("connection not established yet")

// PluginWrapper is the unified wrapper for the interface client of a plugin, it contains metadata of a corresponding
// plugin, buffers recent data samples measured from the corresponding plugin.
type PluginWrapper struct {
	version        string
	lockService    sync.RWMutex
	lockSamples    sync.RWMutex
	samples        map[string]map[int64]types.Price
	client         *plugin.Client
	clientProtocol plugin.ClientProtocol
	name           string
	startAt        time.Time
	logger         hclog.Logger
}

func NewPluginWrapper(name string, pluginDir string) *PluginWrapper {
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

	// We're a host! Start by launching the plugin process.
	rpcClient := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(fmt.Sprintf("%s/%s", pluginDir, name)), //nolint
		Logger:          logger,
	})

	return &PluginWrapper{
		name:    name,
		client:  rpcClient,
		startAt: time.Now(),
		samples: make(map[string]map[int64]types.Price),
		logger:  logger,
	}
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

func (pw *PluginWrapper) Initialize() {
	// connect to remote rpc plugin
	rpcClient, err := pw.client.Client()
	if err != nil {
		pw.logger.Error("cannot connect remote plugin", err.Error())
		return
	}
	pw.clientProtocol = rpcClient
	// load plugin's version
	version, err := pw.GetVersion()
	if err != nil {
		pw.logger.Error("cannot get plugin's version")
		return
	}
	pw.logger.Info("plugin initialized", pw.name, version)
}

func (pw *PluginWrapper) GetVersion() (string, error) {
	if pw.clientProtocol == nil {
		// try to reconnect during the runtime.
		err := pw.connect()
		if err != nil {
			return "", err
		}
	}
	err := pw.clientProtocol.Ping()
	if err != nil {
		pw.clientProtocol.Close() // no lint
		pw.clientProtocol = nil
		// try to reconnect during the runtime.
		err = pw.connect()
		if err != nil {
			return "", err
		}
	}

	raw, err := pw.clientProtocol.Dispense("adapter")
	if err != nil {
		return "", err
	}

	adapter := raw.(types.Adapter)
	pw.version, err = adapter.GetVersion()
	if err != nil {
		return pw.version, err
	}

	return pw.version, nil
}

func (pw *PluginWrapper) FetchPrices(symbols []string, ts int64) error {
	// prevent race condition throughout data sampling routines in case of waiting for timeout.
	pw.lockService.Lock()
	defer pw.lockService.Unlock()

	if pw.clientProtocol == nil {
		// try to reconnect during the runtime.
		err := pw.connect()
		if err != nil {
			pw.logger.Error("connect to plugin", "error", err.Error())
			return err
		}
	}
	err := pw.clientProtocol.Ping()
	if err != nil {
		pw.clientProtocol.Close() // no lint
		pw.clientProtocol = nil
		// try to reconnect during the runtime.
		err = pw.connect()
		if err != nil {
			pw.logger.Error("connect to plugin", "error", err.Error())
			return err
		}
	}

	raw, err := pw.clientProtocol.Dispense("adapter")
	if err != nil {
		pw.logger.Error("Dispense a plugin", "error", err.Error())
		return err
	}

	adapter := raw.(types.Adapter)
	report, err := adapter.FetchPrices(symbols)
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
	pw.client.Kill()
	if pw.clientProtocol != nil {
		pw.clientProtocol.Close() // no lint
	}
}

func (pw *PluginWrapper) connect() error {
	// connect to remote rpc plugin
	rpcClient, err := pw.client.Client()
	if err != nil {
		pw.logger.Error("cannot connect remote plugin", err.Error())
		return errConnectionNotEstablished
	}
	pw.clientProtocol = rpcClient
	return nil
}
