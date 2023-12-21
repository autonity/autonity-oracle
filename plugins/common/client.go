package common

import (
	"net/http"
	"net/url"
	"time"
)

type Connection interface {
	Request(scheme string, endpoint *url.URL) (*http.Response, error)
	Close()
}

type DataSourceClient interface {
	AvailableSymbols() ([]string, error)
	FetchPrice([]string) (Prices, error)
	KeyRequired() bool
	Close()
}

type connection struct {
	client *http.Client
	host   string
}

func NewConnection(duration time.Duration, host string) Connection {
	client := &http.Client{
		Timeout: duration,
	}

	return &connection{
		client: client,
		host:   host,
	}
}
func (conn *connection) Close() {
	if conn.client != nil {
		conn.client.CloseIdleConnections()
	}
}
func (conn *connection) Request(scheme string, endpoint *url.URL) (*http.Response, error) {
	endpoint.Scheme = scheme
	endpoint.Host = conn.host
	targetUrl := endpoint.String()
	return conn.client.Get(targetUrl)
}

type Client struct {
	Conn   Connection
	ApiKey string
}

func NewClient(apiKey string, timeOut time.Duration, host string) *Client {
	return NewClientConnection(apiKey, NewConnection(timeOut, host))
}

func NewClientConnection(apiKey string, connection Connection) *Client {
	return &Client{
		Conn:   connection,
		ApiKey: apiKey,
	}
}
