package pluginclient

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

type PluginClient struct {
	version        string
	pricePool      types.PricePool
	client         *plugin.Client
	clientProtocol plugin.ClientProtocol
	name           string
	startAt        time.Time
	logger         hclog.Logger
}

func NewPluginClient(name string, pluginDir string, pricePool types.PricePool) *PluginClient {
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

	return &PluginClient{
		name:      name,
		client:    rpcClient,
		startAt:   time.Now(),
		pricePool: pricePool,
		logger:    logger,
	}
}

func (ba *PluginClient) Name() string {
	return ba.name
}

func (ba *PluginClient) Version() string {
	return ba.version
}

func (ba *PluginClient) StartTime() time.Time {
	return ba.startAt
}

func (ba *PluginClient) Initialize() {
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

func (ba *PluginClient) GetVersion() (string, error) {
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

func (ba *PluginClient) FetchPrices(symbols []string) error {
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
	prices, err := adapter.FetchPrices(symbols)
	if err != nil {
		return err
	}

	if len(prices) > 0 {
		ba.pricePool.AddPrices(prices)
	}
	return nil
}

func (ba *PluginClient) Close() {
	ba.client.Kill()
	if ba.clientProtocol != nil {
		ba.clientProtocol.Close() // no lint
	}
}

func (ba *PluginClient) connect() error {
	// connect to remote rpc plugin
	rpcClient, err := ba.client.Client()
	if err != nil {
		ba.logger.Error("cannot connect remote plugin", err.Error())
		return errConnectionNotEstablished
	}
	ba.clientProtocol = rpcClient
	return nil
}
