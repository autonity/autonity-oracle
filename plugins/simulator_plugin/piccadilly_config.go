//go:build pic

package main

func init() {
	// set to piccadilly's feed service, to be enabled by build flag: pic
	defaultEndpoint = "simfeed.piccadilly.autonity.org"
}
