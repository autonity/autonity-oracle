package pluginclient

import (
	"autonity-oracle/types"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"log"
	"os"
	"os/exec"
)

var errConnectionNotEstablished = errors.New("connection not established yet")

type PluginClient struct {
	pricePool      types.PricePool
	client         *plugin.Client
	clientProtocol plugin.ClientProtocol
	name           string
}

func NewPluginClient(name string, pluginDir string) *PluginClient {
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
		Cmd:             exec.Command(fmt.Sprintf("%s/%s", pluginDir, name)),
		Logger:          logger,
	})

	return &PluginClient{
		name:   name,
		client: rpcClient,
	}
}

func (ba *PluginClient) Name() string {
	return ba.name
}

func (ba *PluginClient) Initialize(pricePool types.PricePool) {
	ba.pricePool = pricePool

	// connect to remote rpc plugin
	rpcClient, err := ba.client.Client()
	if err != nil {
		log.Printf("cannot connect remote plugin: %s", err.Error())
		return
	}
	ba.clientProtocol = rpcClient
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
		err := ba.connect()
		if err != nil {
			return err
		}
	}

	raw, err := ba.clientProtocol.Dispense("adapter")
	if err != nil {
		return err
	}

	adapter := raw.(types.Adapter)
	prices := adapter.FetchPrices(symbols)

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
		log.Printf("cannot connect remote plugin: %s", err.Error())
		return errConnectionNotEstablished
	}
	ba.clientProtocol = rpcClient
	return nil
}
