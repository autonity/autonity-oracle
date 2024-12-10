//go:build pic

package main

func init() {
	// overwrite the default feed with the piccadilly's feed service, to be enabled by build flag: pic
	defaultEndpoint = "simfeed.piccadilly.autonity.org"
}
