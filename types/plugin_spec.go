package types

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

// This file defines the autonity oracle plugins specification on top of go-plugin framework which leverage the localhost
// net rpc, or grpc for plugin integrations.

// HandshakeConfig are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user-friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginPriceReport is the returned data samples from adapters which carry the prices and those symbols of no data if
// there are any unrecognisable symbols from the data source side.
type PluginPriceReport struct {
	Prices                []Price
	UnRecognizableSymbols []string
}

// PluginState is the returned data when the oracle host want to initialise the plugin with basic information: version,
// and available symbols that the data source support.
type PluginState struct {
	Version          string
	AvailableSymbols []string
}

// Adapter is the interface that we're exposing as a plugin.
type Adapter interface {
	// FetchPrices is called by oracle server to fetch data points of symbols required by the protocol contract, some
	// symbols in the protocol's symbols set might not be recognisable by a data source, thus in the plugin implementation,
	// one need to filter invalid symbols to make sure valid symbol's data be collected. The return PluginPriceReport
	// contains the valid symbols' prices and a set of invalid symbols.
	FetchPrices(symbols []string) (PluginPriceReport, error)
	// State is called by oracle server to instantiate a plugin, it returns the plugin's version and a set of symbol that
	// the data source support, this information is just used for logging at the bootstrap phase of a plugin.
	State() (PluginState, error)
}

// AdapterRPCClient is an implementation that talks over RPC client
type AdapterRPCClient struct{ client *rpc.Client }

func (g *AdapterRPCClient) FetchPrices(symbols []string) (PluginPriceReport, error) {
	var resp PluginPriceReport
	err := g.client.Call("Plugin.FetchPrices", symbols, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (g *AdapterRPCClient) State() (PluginState, error) {
	var resp PluginState
	err := g.client.Call("Plugin.State", new(interface{}), &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// AdapterRPCServer Here is the RPC server that AdapterRPCClient talks to, conforming to the requirements of net/rpc
type AdapterRPCServer struct {
	// This is the real implementation
	Impl Adapter
}

func (s *AdapterRPCServer) FetchPrices(symbols []string, resp *PluginPriceReport) error {
	report, err := s.Impl.FetchPrices(symbols)
	*resp = report
	return err
}

func (s *AdapterRPCServer) State(args interface{}, resp *PluginState) error {
	v, err := s.Impl.State()
	*resp = v
	return err
}

// AdapterPlugin is the unified implementation of plugins, all the 3rd parties plugins need to inject their
// implementation by using this structure in their source code.
type AdapterPlugin struct {
	// Impl Injection
	Impl Adapter
}

func (p *AdapterPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &AdapterRPCServer{Impl: p.Impl}, nil
}

func (AdapterPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &AdapterRPCClient{client: c}, nil
}
