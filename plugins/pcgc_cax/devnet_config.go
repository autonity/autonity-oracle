//go:build dev

package main

func init() {
	// default setting for dev-network, to be enabled by build flag: DVE_NET
	routers = "orderbooks"
	defaultEndpoint = "cax.devnet.clearmatics.network"
}
