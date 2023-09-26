
# Autonity Oracle Server

## Quickstart

With the appropriate parameters, run:

``` 
$./autoracle -oracle_key_file="./keystore/key" -oracle_key_password="123" -oracle_autonity_ws_url="ws://127.0.0.1:8546" -oracle_plugin_conf="./plugins/plugin-conf.yml"
```  

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
-  The validator must participate in the Oracle protocol if selected to be part of the consensus committee.
- A proof of ownership for the oracle account is required when submitting a registration transaction to the Autonity Contract.
- Oracle transactions are refunded if successful. The balance of the oracle account is required to be funded to cover at least one voting transaction.

## Data adaptors - plugin architecture
The Autonity oracle client provides a unified interface between itself and plugins that adapt data from  
different data providers, for example Binance and Coingecko, etc. Any party can build a plugin implementing  
this unified interface and so provide an adaptor for any data source on demand. The oracle client will scan and load  
plugins from the plugin directory during runtime. Detection of new or changed plugins is dynamic;  
no shutdown of the oracle client is required to detect and apply the change.

## Coordination of data sampling
To coordinate data sampling in the oracle network, the L1 oracle contract issues a round event on every vote period (60 blocks). The round event carries a tuple `(RoundID, SampleTS, Height, VotePeriod)`, which tell the oracle clients that on round with ID `RoundID`, a data sample with timestamp `SampleTS` is required for the data submission. The `Height` stands for the start height of the new round, while the `VotePeriod` stands for the round length of the new round. Thus the oracle client can estimate and manage data pre-samplings for the new round and then pick up the nearest sample refering to the required `SampleTS`.

![Screenshot from 2023-04-21 04-19-10](https://user-images.githubusercontent.com/54585152/233533092-29b65a39-eb87-496f-9a1e-0741bc7fbd45.png)

## Configuration
Values that can be configured by using environment variables:

| **Env Variable** | **Required?** | **Meaning** | **Default Value**                                                                                    | **Valid Options** |
|----------------------------|---------------|---------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| `ORACLE_SYMBOLS` | No | The symbols that the oracle component collects data points for | "AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN"                                    | symbols separated by ',' |
| `ORACLE_PLUGIN_DIR` | No | The directory that stores the plugins | "./build/bin/plugins"                                                                                | any directory that saves plugins |
| `ORACLE_KEY_FILE` | Yes | The encrypted key file path that contains the private key of the oracle client. | "./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" | any key file that saves the private key |  
| `ORACLE_KEY_PASSWORD` | Yes | The password of the key file that contains the private key of the oracle client. | "123"                                                                                                | any password that encrypted the private key |
| `ORACLE_AUTONITY_WS_URL` | Yes | The web socket RPC URL of your Autonity L1 Node that the oracle client communicates with. | "ws://127.0.0.1:8546"                                                                                | the web socket rpc endpoint url of the Autonity client. |
| `ORACLE_PLUGIN_CONF` | Yes | The plugins' configuration file in YAML. | "./build/bin/plugins/plugins-conf.yml"                                                               | the configuration file of the oracle plugins. |
| `ORACLE_GAS_TIP_CAP` | No | The gas priority fee cap to issue the oracle data report transactions | 1                                                               | A non-zero value per gas to prioritize your data report TX to be mined. |


or by using console flags:
```shell
$./autoracle -help  
Usage of ./autoracle:
  -oracle_autonity_ws_url="ws://127.0.0.1:8546": WS-RPC server listening interface and port of the connected Autonity Go Client node
  -oracle_gas_tip_cap=1: The gas priority fee cap set for oracle data report transactions
  -oracle_key_file="./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe": Oracle server key file
  -oracle_key_password="123": Password to the oracle server key file
  -oracle_plugin_conf="./build/bin/plugins/plugins-conf.yml": The plugins' configuration file in YAML
  -oracle_plugin_dir="./build/bin/plugins": The DIR where the adapter plugins are stored
  -oracle_symbols="AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN": The currency pair symbols the oracle returns data for. A comma-separated list
```


example to run the autonity oracle service with console flags:
```shell
$./autoracle -oracle_symbols="AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN" -oracle_plugin_dir="./plugins" -oracle_key_file="../../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" -oracle_key_password="123" -oracle_autonity_ws_url="ws://127.0.0.1:8546" -oracle_plugin_conf="./plugins/plugins-conf.yml"
```
plugin configuration file:    

`A yaml file to config extra data for a specific plugin nominated by its name, for example here is a plugin configuration to
config the service key to the corresponding plugins:`

```yaml
# A list of plugins with each one's executable binary name and its corresponding API key granted by the data provider.
# For forex data providers which provide the realtime exchange rate of EUR/USD, JPY/USD, GBP/USD, AUD/USD, CAD/USD and
# SEK/USD, there are 4 official plugins listed below, end user can select one or more of them on-demand for the forex
# data sourcing. Remember to config the API key in below once you subscribe the service plan from your data vendors.

# Un-comment below lines to enable your configuration on demand.
#  - name: forex_currencyfreaks              # visit https://currencyfreaks.com to apply for your key
#    key: 5490e15565e741129788f6100e022ec5   # replace it with your own key

#  - name: forex_openexchangerate            # visit https://openexchangerates.org to apply for your key
#    key: 0be02ca33c4843ee968c4cedd2686f01   # replace it with your own key

#  - name: forex_currencylayer               # visit https://currencylayer.com to apply for your key
#    key: 705af082ac7f7d150c87303d4e2f049e   # replace it with your own key

#  - name: forex_exchangerate                # visit https://www.exchangerate-api.com to apply for your key
#    key: 411f04e4775bb86c20296530           # replace it with your own key

- name: forex_currencyfreaks              # visit https://currencyfreaks.com to apply for your key
  key: 5490e15565e741129788f6100e022ec5   # replace it with your own key

- name: forex_openexchangerate            # visit https://openexchangerates.org to apply for your key
  key: 0be02ca33c4843ee968c4cedd2686f01   # replace it with your own key

- name: forex_currencylayer               # visit https://currencylayer.com to apply for your key
  key: 705af082ac7f7d150c87303d4e2f049e   # replace it with your own key

- name: forex_exchangerate                # visit https://www.exchangerate-api.com to apply for your key
  key: 411f04e4775bb86c20296530           # replace it with your own key
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

$export ORACLE_SYMBOLS="AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN"
$export ORACLE_PLUGIN_DIR="./plugins"  
$export ORACLE_KEY_FILE="./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"  
$export ORACLE_KEY_PASSWORD="your passord to the key file"  
$export ORACLE_AUTONITY_WS_URL="ws://127.0.0.1:8546"
$export ORACLE_PLUGIN_CONF="./plugins/plugins-conf.yml"
$.~/src/autonity-oracle/build/bin/autoracle
```

or configure by using console flags and run the binary:

```shell
$./autoracle -oracle_symbols="AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN" -oracle_plugin_dir="./plugins" -oracle_key_file="../../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" -oracle_key_password="123" -oracle_autonity_ws_url="ws://127.0.0.1:8546"
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
ExecStart=/home/test/src/autonity-oracle/build/bin/autoracle -oracle_plugin_dir="/home/test/src/autonity-oracle/build/bin/plugins" -oracle_plugin_conf="/home/test/src/autonity-oracle/build/bin/plugins/plugins-conf.yml"
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
  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th systemd[1]: Stopping Clearmatics Autonity Oracle Server...  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.233Z [INFO] *oracleserver.OracleServer: the jobTicker jobs of oracle service is stopped  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.233Z [DEBUG] binance.binance: 2023/01/19 03:03:45 [DEBUG] plugin: plugin server: accept unix /tmp/plugin3024381010: use of closed network connection  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.235Z [INFO] binance: plugin process exited: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance pid=14577  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.235Z [DEBUG] binance: plugin exited  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.236Z [DEBUG] fakeplugin.fakeplugin: 2023/01/19 03:03:45 [DEBUG] plugin: plugin server: accept unix /tmp/plugin2424636505: use of closed network connection  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.237Z [INFO] fakeplugin: plugin process exited: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin pid=14586  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.237Z [DEBUG] fakeplugin: plugin exited  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th systemd[1]: autoracle.service: Succeeded.  
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th systemd[1]: Stopped Clearmatics Autonity Oracle Server.  
  
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
├─14568 /home/jason/src/autonity-oracle/build/bin/autoracle -oracle_plugin_dir=/home/jason/src/autonity-oracle/build/bin/plugins  
├─14577 /home/jason/src/autonity-oracle/build/bin/plugins/binance  
└─14586 /home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin  
  
Jan 19 02:57:39 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:57:39.155Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:57:39.154Z  
Jan 19 02:57:59 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:57:59.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:57:59.156Z  
  
```  

#### Collect system logs

sudo journalctl -u autoracle.service -b

```  
-- Logs begin at Sat 2022-11-26 11:54:00 GMT, end at Thu 2023-01-19 02:59:51 GMT. --  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th systemd[1]: Started Clearmatics Autonity Oracle Server.  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023/01/19 02:42:19  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: Running autonity oracle service at port: 30311, with symbols: NTNUSDT,NTNUSDC,NTNBTC,NTNETH and plugin diretory: /home/jason/src/autonity-oracle/build/bin/plugins  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [WARN] binance: plugin configured with a nil SecureConfig  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: starting plugin: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance args=[/home/jason/src/autonity-oracle/build/bin/plugins/binance]  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: plugin started: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance pid=14577  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: waiting for RPC address: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.159Z [DEBUG] binance.binance: plugin address: network=unix address=/tmp/plugin3024381010 timestamp=2023-01-19T02:42:19.159Z  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.159Z [DEBUG] binance: using plugin: version=1  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [INFO] binance: plugin initialized: binance=v0.0.1  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [WARN] fakeplugin: plugin configured with a nil SecureConfig  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: starting plugin: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin args=[/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin]  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: plugin started: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin pid=14586  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: waiting for RPC address: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.168Z [DEBUG] fakeplugin.fakeplugin: plugin address: address=/tmp/plugin2424636505 network=unix timestamp=2023-01-19T02:42:19.167Z  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.168Z [DEBUG] fakeplugin: using plugin: version=1  
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.170Z [INFO] fakeplugin: plugin initialized: fakeplugin=v0.0.1  
Jan 19 02:42:29 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:29.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:42:29.156Z  
Jan 19 02:43:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:43:19.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:43:19.156Z  
```  

### Runtime plugin management
#### Adding new plugins
To add a new data source, just put the new plugin into the service's `plugins` directory. The oracle service auto discovers and manages it. There are no other operations required from the operator.
#### Replace running plugins
To replace running plugins with new ones, just replace the binary in the `plugins` directory. The oracle service auto discovers it by checking the modification time of the binary and does the plugin replacement itself. There are no other operations required from the operator.


## Development

To build the project run
```shell
make autoracle
```
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

