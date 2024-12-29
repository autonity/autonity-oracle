// Copyright 2021 The go-ethereum Authors
// This file is part of go-ethereum.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package metrics

// Config contains the configuration for the metric collection.
type Config struct {
	// Common configs for influxDB V1 and V2.
	InfluxDBEndpoint string `json:"influxDBEndpoint" yaml:"influxDBEndpoint"`
	InfluxDBTags     string `json:"influxDBTags" yaml:"influxDBTags"`

	// InfluxDB V1 specific configs
	EnableInfluxDB   bool   `json:"enableInfluxDB" yaml:"enableInfluxDB"`
	InfluxDBDatabase string `json:"influxDBDatabase" yaml:"influxDBDatabase"`
	InfluxDBUsername string `json:"influxDBUsername" yaml:"influxDBUsername"`
	InfluxDBPassword string `json:"influxDBPassword" yaml:"influxDBPassword"`

	// InfluxDB V2 specific configs
	EnableInfluxDBV2     bool   `json:"enableInfluxDBV2" yaml:"enableInfluxDBV2"`
	InfluxDBToken        string `json:"influxDBToken" yaml:"influxDBToken"`
	InfluxDBBucket       string `json:"influxDBBucket" yaml:"influxDBBucket"`
	InfluxDBOrganization string `json:"influxDBOrganization" yaml:"influxDBOrganization"`
}

// DefaultConfig is the default config for metrics used in go-ethereum.
var DefaultConfig = Config{
	// common flags
	InfluxDBEndpoint: "http://localhost:8086",
	InfluxDBTags:     "host=localhost",

	// influxdbv1-specific flags.
	EnableInfluxDB:   false,
	InfluxDBDatabase: "autonity",
	InfluxDBUsername: "test",
	InfluxDBPassword: "test",

	// influxdbv2-specific flags
	EnableInfluxDBV2:     false,
	InfluxDBToken:        "test",
	InfluxDBBucket:       "autonity",
	InfluxDBOrganization: "autonity",
}
