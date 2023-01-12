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

// Adapter is the interface that we're exposing as a plugin.
type Adapter interface {
	FetchPrices(symbols []string) ([]Price, error)
	GetVersion() (string, error)
}

// AdapterRPCClient is an implementation that talks over RPC client
type AdapterRPCClient struct{ client *rpc.Client }

func (g *AdapterRPCClient) FetchPrices(symbols []string) ([]Price, error) {
	var resp []Price
	err := g.client.Call("Plugin.FetchPrices", symbols, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *AdapterRPCClient) GetVersion() (string, error) {
	var resp string
	err := g.client.Call("Plugin.GetVersion", new(interface{}), &resp)
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

func (s *AdapterRPCServer) FetchPrices(symbols []string, resp *[]Price) error {
	prices, err := s.Impl.FetchPrices(symbols)
	*resp = prices
	return err
}

func (s *AdapterRPCServer) GetVersion(args interface{}, resp *string) error {
	v, err := s.Impl.GetVersion()
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
