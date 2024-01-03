
# Autonity Oracle Server

## Assumptions

This project assumes the following:

* Go 1.19.3
* Linux operating system

## Overview

The Autonity Oracle Server offers exchange rate data for currency pairs from various data providers,  
consolidating these data points into a standardized report. This information is then pushed to the  
oracle contract deployed on the Autonity L1 network via a transaction. The oracle contract aggregates those supplied data   points to supply a reference exchange rate data through a voting mechanism using a commit-reveal scheme.  
DApps deployed on the Autonity L1 network can access these data points via the oracle contract interface.

## The Oracle Server
- The Oracle Server is operated and maintained by an Autonity validator node operator.
- The validator must participate in the Oracle protocol if selected to be part of the consensus committee.
- A proof of ownership for the oracle account is required when submitting a registration transaction to the Autonity Contract.
- Oracle transactions are refunded if successful. The balance of the oracle account is required to be funded to cover at least one voting transaction.

## Data adaptors - plugin architecture
The Autonity oracle client provides a unified interface between itself and plugins that adapt data from  
different data providers, for example Binance and Coingecko, etc. Any party can build a plugin implementing  
this unified interface and so provide an adaptor for any data source on demand. The oracle client will scan and load  
plugins from the plugin directory during runtime. Detection of new or changed plugins is dynamic;  
no shutdown of the oracle client is required to detect and apply the change.

## Coordination of data sampling
### Overview
To coordinate data sampling in the oracle network, the L1 oracle contract issues a round event on every vote period (30 ~ 60 blocks). The round event carries a tuple `(RoundID, SampleTS, Height, VotePeriod)`, which tells the oracle servers that on round with ID `RoundID`, a data sample with timestamp `SampleTS` is required for the data submission. The `Height` stands for the start height of the new round, while the `VotePeriod` stands for the round length of the new round. Thus the oracle server can estimate and manage data pre-samplings for the new round and then pick up the nearest sample referring to the required `SampleTS`.

![Screenshot from 2023-04-21 04-19-10](https://user-images.githubusercontent.com/54585152/233533092-29b65a39-eb87-496f-9a1e-0741bc7fbd45.png)
### The data pre-sampling
To mitigate data deviation caused by the distributed system environment, a data pre-sampling mechanism is employed parameterised by `SampleTS` and `Height` log data from the round event. When approaching the next round's start boundary `Height`, the oracle server initiates data pre-sampling approximately 15 seconds in advance. During this pre-sampling window, the server samples data per second and selects the sample closest to the required `SampleTS` for data aggregation. The oracle server will then submit that sample to the L1 oracle contract as its price vote for the next oracle voting round.    

In a production network, node operators should obtain real-time data from high-quality data sources. However, most commercial data providers price their services based on quality of service (QoS) and rate limits. To address this, a configuration parameter "refresh" has been introduced for each data plugin. This parameter represents the interval in seconds between data fetches after the last successful data sampling. A buffered sample is used before the next data fetch. Node operators should configure an appropriate "refresh" interval by estimating the data fetching rate and the QoS subscribed from the data provider. The default value of "refresh" is 30 seconds, indicating that the plugin will query the data from the data source once every 30 seconds, even during the data pre-sampling window. If the data source does not limit the rate, it's recommended to set "refresh" to 1, allowing the pre-sampling to fetch data every 1 second to obtain real-time data. If the default "refresh" of 30 seconds is kept, then the oracle server will be sampling data up to 30 seconds old rather than in real-time.

## Version

Print the version of the oracle server:

```
$./autoracle version
v0.1.3
```

## Configuration
### Oracle Server Config
Values that can be configured by using environment variables:

| **Env Variable** | **Required?** | **Meaning** | **Default Value**                                                                                    | **Valid Options** |
|----------------------------|---------------|---------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| `PLUGIN_DIR` | No | The directory that stores the plugins | "./plugins"                                                                                | any directory that saves plugins |
| `KEY_FILE` | Yes | The encrypted key file path that contains the private key of the oracle client. | "./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" | any key file that saves the private key |
| `KEY_PASSWORD` | Yes | The password of the key file that contains the private key of the oracle client. | "123"                                                                                                | any password that encrypted the private key |
| `AUTONITY_WS` | Yes | The web socket RPC URL of your Autonity L1 Node that the oracle client communicates with. | "ws://127.0.0.1:8546"                                                                                | the web socket rpc endpoint url of the Autonity client. |
| `PLUGIN_CONF` | Yes | The plugins' configuration file in YAML. | "./plugins-conf.yml"                                                               | the configuration file of the oracle plugins. |
| `GAS_TIP_CAP` | No | The gas priority fee cap to issue the oracle data report transactions | 1                                                               | A non-zero value per gas to prioritize your data report TX to be mined. |
| `LOG_LEVEL` | No | The logging level of the oracle server | 3                                                              | available levels are:  0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error. |


or by using console flags:
```shell
$./autoracle --help
Usage of Autonity Oracle Server:
Sub commands: 
  version: print the version of the oracle server.
Flags:
  -key.file="./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe": Set oracle server key file
  -key.password="123": Set the password to decrypt oracle server key file
  -log.level=2: Set the logging level, available levels are:  0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error
  -plugin.conf="./plugins-conf.yml": Set the plugins' configuration file
  -plugin.dir="./plugins": Set the directory of the data plugins.
  -tip=1: Set the gas priority fee cap to issue the oracle data report transactions.
  -ws="ws://127.0.0.1:8546": Set the WS-RPC server listening interface and port of the connected Autonity Client node

```


example to run the autonity oracle service with console flags:
```shell
$./autoracle --plugin.dir="./plugins" --key.file="../../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" --key.password="123" --ws="ws://127.0.0.1:8546" --plugin.conf="./plugins-conf.yml"
```
### Plugin Config    

`A yaml file to configure plugins:`

```yaml
# The forex data plugins are used to fetch realtime rate of currency pairs:
# EUR-USD, JPY-USD, GBP-USD, AUD-USD, CAD-USD and SEK-USD from commercial data providers.
# There are 4 implemented forex data plugins, each of them requires the end user to apply for their own service key from
# the selected data provider. The selection of the forex data plugin is on demand by the end user. The user can use anyone
# of them, or he/she can use multiple forex data plugins in the setup.
#
# The crypto data plugins are used to fetch realtime rate of crypto currency pairs:
# ATN-USD, NTN-USD, NTN-ATN from exchanges. For the Autonity Piccadilly Circus Games Competition Round 4 game, the data provider of these pairs is a simulated
# exchange that people can trade ATN and NTN in markets created by it. No configuration is required for
# the plugin named pcgc_cax that fetches data for these crypto currency pairs from this exchange as the default configuration is set in the plugin.

# For the forex data plugin, default
# configuration is set, so the end user just needs to configure required settings, namely `name` and `key`. The configuration settings of a plugin are:
#
# type PluginConfig struct {
#	Name               string `json:"name" yaml:"name"`         // the name of the plugin binary, it is required.
#	Key                string `json:"key" yaml:"key"`           // the API key granted by data provider, it is required.
#	Scheme             string `json:"scheme" yaml:"scheme"`     // the data service scheme, http or https, it is optional.
#	Endpoint           string `json:"endpoint" yaml:"endpoint"` // the hostname of the data service endpoint, it is optional.
#	Timeout            int    `json:"timeout" yaml:"timeout"`   // the timeout in seconds that a request last for, it is optional.
#	DataUpdateInterval int    `json:"refresh" yaml:"refresh"`   // the interval in seconds to fetch data due to the rate limit from the provider.
#}

# As an example, to set the configuration of the plugin `forex_currencyfreaks`, only the required field are needed,
# however you can configure the optional fields on demand to fit your service quality from the rate provider.
#  - name: forex_currencyfreaks              # required, it is the plugin file name in the plugin directory.
#    key: 575aab9e47e54790bf6d502c48407c10   # required, visit https://currencyfreaks.com to get your key, and replace it.
#    scheme: https                           # optional, https or http, default value is https.
#    endpoint: api.currencyfreaks.com        # optional, default value is api.currencyfreaks.com
#    timeout: 10                             # optional, default value is 10.
#    refresh: 30                             # optional, default value is 30, that is 30 seconds to fetch data from data source.

# Un-comment below lines to enable your forex data plugin's configuration on demand. Your production configurations starts from below:

#  - name: forex_currencyfreaks              # required, it is the plugin file name in the plugin directory.
#    key: 575aab9e47e54790bf6d502c48407c10   # required, visit https://currencyfreaks.com to get your key, and replace it.

#  - name: forex_openexchangerate            # required, it is the plugin file name in the plugin directory.
#    key: 0be02ca33c4843ee968c4cedd2686f01   # required, visit https://openexchangerates.org to get your key, and replace it.

#  - name: forex_currencylayer               # required, it is the plugin file name in the plugin directory.
#    key: 705af082ac7f7d150c87303d4e2f049e   # required, visit https://currencylayer.com  to get your key, and replace it.

#  - name: forex_exchangerate                # required, it is the plugin file name in the plugin directory.
#    key: 411f04e4775bb86c20296530           # required, visit https://www.exchangerate-api.com to get your key, and replace it.

```

Available configuration fields:

There are multiple configuration fields can be used, it is not required to config each field of them, it depends on your plugin implementation.
```go
// PluginConfig carry the configuration of plugins.
type PluginConfig struct {
	Name               string `json:"name" yaml:"name"`         // the name of the plugin binary.
	Key                string `json:"key" yaml:"key"`           // the API key granted by your data provider to access their data API.
	Scheme             string `json:"scheme" yaml:"scheme"`     // the data service scheme, http or https.
	Endpoint           string `json:"endpoint" yaml:"endpoint"` // the data service endpoint url of the data provider.
	Timeout            int    `json:"timeout" yaml:"timeout"`   // the timeout period that an API request is lasting for.
	DataUpdateInterval int    `json:"refresh" yaml:"refresh"`   // reserved for rate limited provider's plugin, limit the request rate.
}
```
In the last configuration file, all the forex data vendors need a service key to access their data, thus a key is expected for the corresponding plugins.


## Deployment
### Oracle Client Private Key generation
Download the Autonity client to generate the private key from console, and set the password to encode the key file, the  
key file path will display, and remember the password that encrypted the key file.

```shell
$./autonity --datadir ./keys/ account new  
Your new account is locked with a password. Please give a password. Do not forget this password.  
Password:xxxxxx  
Repeat password:xxxxxx

Your new key was generated

Public address of the key: 0x7C785Fe9404574AaC7daf2FF30637546493900d1  
Path of the secret key file: key-data/keystore/UTC--2023-02-28T11-40-15.383709761Z--7c785fe9404574aac7daf2ff30637546493900d1

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
```

### Start up the service from shell console
Prepare the plugin binaries, and save them into the `plugins` directory. To start the service, set the system environment variables and run the binary:
```shell

$export SYMBOLS="AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN"
$export PLUGIN_DIR="./plugins"
$export KEY_FILE="./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
$export KEY_PASSWORD="your passord to the key file"
$export AUTONITY_WS="ws://127.0.0.1:8546"
$export PLUGIN_CONF="./plugins-conf.yml"
$.~/src/autonity-oracle/build/bin/autoracle
```

or configure by using console flags and run the binary:

```shell
$./autoracle --plugin.dir="./plugins" --key.file="./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" --key.password="123" --ws="ws://127.0.0.1:8546"
```

### Deploy via linux system daemon
#### Preparations
Prepare the configurations via system environment variables and the corresponding plugin binaries. Create a service registration file under your service discovery DIR of the system daemon, for example "/etc/systemd/system/" in Ubuntu Linux.  
Here I create a service registration file called "/etc/systemd/system/autoracle.service" with content:
```  
[Unit]  
Description=Clearmatics Autonity Oracle Server  
After=syslog.target network.target  
[Service]  
Type=simple  
ExecStart=/home/test/src/autonity-oracle/build/bin/autoracle -plugin.dir="/home/test/src/autonity-oracle/build/bin/plugins" -plugin.conf="/home/test/src/autonity-oracle/build/bin/plugins/plugins-conf.yml"
KillMode=process  
KillSignal=SIGINT  
TimeoutStopSec=5  
Restart=on-failure  
RestartSec=5  
[Install]  
Alias=autoracle.service  
WantedBy=multi-user.target  
```  
#### Start the service

```shell
sudo systemctl start autoracle.service
```

#### Stop the service
```shell
sudo systemctl stop autoracle.service
● autoracle.service - Clearmatics Autonity Oracle Server  
Loaded: loaded (/etc/systemd/system/autoracle.service; disabled; vendor preset: enabled)  
Active: inactive (dead)  
  
Jan 19 03:03:45 systemd[1]: Stopping Clearmatics Autonity Oracle Server...  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.233Z [INFO] *oracleserver.OracleServer: the jobTicker jobs of oracle service is stopped  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.233Z [DEBUG] binance.binance: 2023/01/19 03:03:45 [DEBUG] plugin: plugin server: accept unix /tmp/plugin3024381010: use of closed network connection  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.235Z [INFO] binance: plugin process exited: path=/home/user/src/autonity-oracle/build/bin/plugins/binance pid=14577  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.235Z [DEBUG] binance: plugin exited  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.236Z [DEBUG] fakeplugin.fakeplugin: 2023/01/19 03:03:45 [DEBUG] plugin: plugin server: accept unix /tmp/plugin2424636505: use of closed network connection  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.237Z [INFO] fakeplugin: plugin process exited: path=/home/user/src/autonity-oracle/build/bin/plugins/fakeplugin pid=14586  
Jan 19 03:03:45 autoracle[14568]: 2023-01-19T03:03:45.237Z [DEBUG] fakeplugin: plugin exited  
Jan 19 03:03:45 systemd[1]: autoracle.service: Succeeded.  
Jan 19 03:03:45 systemd[1]: Stopped Clearmatics Autonity Oracle Server.  
  
```  

#### Check the runtime status

```
sudo systemctl status autoracle.service
  
● autoracle.service - Clearmatics Autonity Oracle Server  
Loaded: loaded (/etc/systemd/system/autoracle.service; disabled; vendor preset: enabled)  
Active: active (running) since Thu 2023-01-19 02:42:19 GMT; 15min ago  
Main PID: 14568 (autoracle)  
Tasks: 34 (limit: 18690)  
Memory: 25.4M  
CGroup: /system.slice/autoracle.service  
├─14568 /home/user/src/autonity-oracle/build/bin/autoracle -plugin.dir=/home/user/src/autonity-oracle/build/bin/plugins
├─14577 /home/user/src/autonity-oracle/build/bin/plugins/binance  
└─14586 /home/user/src/autonity-oracle/build/bin/plugins/fakeplugin  
  
Jan 19 02:57:39 autoracle[14568]: 2023-01-19T02:57:39.155Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:57:39.154Z  
Jan 19 02:57:59 autoracle[14568]: 2023-01-19T02:57:59.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:57:59.156Z  
  
```  

#### Collect system logs

sudo journalctl -u autoracle.service -b

```  
-- Logs begin at Sat 2022-11-26 11:54:00 GMT, end at Thu 2023-01-19 02:59:51 GMT. --  
Jan 19 02:42:19 systemd[1]: Started Clearmatics Autonity Oracle Server.  
Jan 19 02:42:19 autoracle[14568]: 2023/01/19 02:42:19  
Jan 19 02:42:19 autoracle[14568]: Running autonity oracle service at port: 30311, with plugin diretory: /home/user/src/autonity-oracle/build/bin/plugins
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.152Z [WARN] binance: plugin configured with a nil SecureConfig  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: starting plugin: path=/home/user/src/autonity-oracle/build/bin/plugins/binance args=[/home/uesr/src/autonity-oracle/build/bin/plugins/binance]  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: plugin started: path=/home/user/src/autonity-oracle/build/bin/plugins/binance pid=14577  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: waiting for RPC address: path=/home/user/src/autonity-oracle/build/bin/plugins/binance  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.159Z [DEBUG] binance.binance: plugin address: network=unix address=/tmp/plugin3024381010 timestamp=2023-01-19T02:42:19.159Z  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.159Z [DEBUG] binance: using plugin: version=1  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.161Z [INFO] binance: plugin initialized: binance=v0.0.1  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.161Z [WARN] fakeplugin: plugin configured with a nil SecureConfig  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: starting plugin: path=/home/user/src/autonity-oracle/build/bin/plugins/fakeplugin args=[/home/user/src/autonity-oracle/build/bin/plugins/fakeplugin]  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: plugin started: path=/home/user/src/autonity-oracle/build/bin/plugins/fakeplugin pid=14586  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: waiting for RPC address: path=/home/user/src/autonity-oracle/build/bin/plugins/fakeplugin  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.168Z [DEBUG] fakeplugin.fakeplugin: plugin address: address=/tmp/plugin2424636505 network=unix timestamp=2023-01-19T02:42:19.167Z  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.168Z [DEBUG] fakeplugin: using plugin: version=1  
Jan 19 02:42:19 autoracle[14568]: 2023-01-19T02:42:19.170Z [INFO] fakeplugin: plugin initialized: fakeplugin=v0.0.1  
Jan 19 02:42:29 autoracle[14568]: 2023-01-19T02:42:29.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:42:29.156Z  
Jan 19 02:43:19 autoracle[14568]: 2023-01-19T02:43:19.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:43:19.156Z  
```  

### Runtime plugin management
#### Adding new plugins
To add a new data source, just put the new plugin into the service's `plugins` directory. The oracle service auto discovers and manages it. There are no other operations required from the operator.
#### Replace running plugins
To replace running plugins with new ones, just replace the binary in the `plugins` directory. The oracle service auto discovers it by checking the modification time of the binary and does the plugin replacement itself. There are no other operations required from the operator.


## Development
### Build for DEV net
```shell
make autoracle-dev
```
### Build for Bakerloo net
```shell
make autoracle-bakerloo
```
### Build for Piccadilly net
```shell
make autoracle
```
### Other build helpers
To build the data source simulator run
```shell
make simulator
```
To run e2e test use
```shell
make e2e-test
```
To run all tests use
```shell
make test
```
To lint code run
```shell
make lint
```
To generate mocks for unit test
```shell
make mock
```
To build docker image
```shell
make build-docker-image
```
To build a plugin, please refer to [How to build a plugin](plugins/README.md)

Built binaries are presented at: `./build/bin` under which there is a `plugins` directory for the built plugins as well.

