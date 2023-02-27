package pluginwrapper

import (
	"autonity-oracle/types"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"
	"time"
)

var errConnectionNotEstablished = errors.New("connection not established yet")

type PluginWrapper struct {
	version        string
	pricePool      types.PricePool
	client         *plugin.Client
	clientProtocol plugin.ClientProtocol
	name           string
	startAt        time.Time
	logger         hclog.Logger
}

func NewPluginWrapper(name string, pluginDir string, pricePool types.PricePool) *PluginWrapper {
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
		name:      name,
		client:    rpcClient,
		startAt:   time.Now(),
		pricePool: pricePool,
		logger:    logger,
	}
}

func (ba *PluginWrapper) Name() string {
	return ba.name
}

func (ba *PluginWrapper) Version() string {
	return ba.version
}

func (ba *PluginWrapper) StartTime() time.Time {
	return ba.startAt
}

func (ba *PluginWrapper) Initialize() {
	// connect to remote rpc plugin
	rpcClient, err := ba.client.Client()
	if err != nil {
		ba.logger.Error("cannot connect remote plugin", err.Error())
		return
	}
	ba.clientProtocol = rpcClient
	// load plugin's version
	version, err := ba.GetVersion()
	if err != nil {
		ba.logger.Warn("cannot get plugin's version")
		return
	}
	ba.logger.Info("plugin initialized", ba.name, version)
}

func (ba *PluginWrapper) GetVersion() (string, error) {
	if ba.clientProtocol == nil {
		// try to reconnect during the runtime.
		err := ba.connect()
		if err != nil {
			return "", err
		}
	}
	err := ba.clientProtocol.Ping()
	if err != nil {
		ba.clientProtocol.Close() // no lint
		ba.clientProtocol = nil
		// try to reconnect during the runtime.
		err = ba.connect()
		if err != nil {
			return "", err
		}
	}

	raw, err := ba.clientProtocol.Dispense("adapter")
	if err != nil {
		return "", err
	}

	adapter := raw.(types.Adapter)
	ba.version, err = adapter.GetVersion()
	if err != nil {
		return ba.version, err
	}

	return ba.version, nil
}

func (ba *PluginWrapper) FetchPrices(symbols []string) error {
	if ba.clientProtocol == nil {
		// try to reconnect during the runtime.
		err := ba.connect()
		if err != nil {
			return err
		}
	}
	err := ba.clientProtocol.Ping()
	if err != nil {
		ba.clientProtocol.Close() // no lint
		ba.clientProtocol = nil
		// try to reconnect during the runtime.
		err = ba.connect()
		if err != nil {
			return err
		}
	}

	raw, err := ba.clientProtocol.Dispense("adapter")
	if err != nil {
		return err
	}

	adapter := raw.(types.Adapter)
	report, err := adapter.FetchPrices(symbols)
	if len(report.BadSymbols) != 0 {
		ba.logger.Warn("find bad symbols: ", report.BadSymbols)
	}
	if err != nil {
		return err
	}

	if len(report.Prices) > 0 {
		ba.pricePool.AddPrices(report.Prices)
	}
	return nil
}

func (ba *PluginWrapper) Close() {
	ba.client.Kill()
	if ba.clientProtocol != nil {
		ba.clientProtocol.Close() // no lint
	}
}

func (ba *PluginWrapper) connect() error {
	// connect to remote rpc plugin
	rpcClient, err := ba.client.Client()
	if err != nil {
		ba.logger.Error("cannot connect remote plugin", err.Error())
		return errConnectionNotEstablished
	}
	ba.clientProtocol = rpcClient
	return nil
}
