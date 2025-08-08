
# Autonity Oracle Server

## Assumptions

This project assumes the following:

* Go 1.24
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
To coordinate data sampling in the oracle network, the L1 oracle contract issues a round event on every vote period. The round event carries a tuple `(RoundID, SampleTS, Height, VotePeriod)`, which tells the oracle servers that on round with ID `RoundID`, a data sample with timestamp `SampleTS` is required for the data submission. The `Height` stands for the start height of the new round, while the `VotePeriod` stands for the round length of the new round. Thus the oracle server can estimate and manage data pre-samplings for the new round and then pick up the nearest sample referring to the required `SampleTS`.

<img width="1564" height="819" alt="Screenshot from 2025-08-07 13-03-30" src="https://github.com/user-attachments/assets/10aa0578-bdd6-4222-aee6-f8abbd37d50d" />



### Data pre-sampling
To mitigate data deviation caused by the distributed system environment, a data pre-sampling mechanism is employed parameterised by `SampleTS` and `Height` log data from the round event. When approaching the next round's start boundary `Height`, the oracle server initiates data pre-sampling approximately 15 seconds in advance. During this pre-sampling window, the server samples data per second and selects the sample closest to the required `SampleTS` for data aggregation. The oracle server will then submit that sample to the L1 oracle contract as its price vote for the next oracle voting round.    

In a production network, node operators should obtain real-time data from high-quality data sources. However, most commercial data providers price their services based on quality of service (QoS) and rate limits. To address this, a configuration parameter "refresh" has been introduced for each data plugin. This parameter represents the interval in seconds between data fetches after the last successful data sampling. A buffered sample is used before the next data fetch. Node operators should configure an appropriate "refresh" interval by estimating the data fetching rate and the QoS subscribed from the data provider. The default value of "refresh" is 30 seconds, indicating that the plugin will query the data from the data source once every 30 seconds, even during the data pre-sampling window. If the data source does not limit the rate, it's recommended to set "refresh" to 1, allowing the pre-sampling to fetch data every 1 second to obtain real-time data. If the default "refresh" of 30 seconds is kept, then the oracle server will be sampling data up to 30 seconds old rather than in real-time.

## Configuration
The configuration file:
```yaml
# Oracle Server Configuration

# Below is the list of default configuration for oracle server:
logLevel: 3  # Logging verbosity: 0: NoLevel, 1: Trace, 2: Debug, 3: Info, 4: Warn, 5: Error
gasTipCap: 1000000000  # 1GWei, the gas priority fee cap for oracle vote message which will be reimbursed by Autonity network.

#Set the buffering time window in blocks to continue vote after the last penalty event. Default value is 86400 (1 day).
#With such time buffer, the node operator can check and repair the local infra without being slashed due to the voting.
#This is important for node operator to prevent node from getting slashed again.
voteBuffer: 86400  # Buffer time in seconds (3600 * 24)

#Set oracle server key file.
keyFile: "./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"

#Set the password to decrypt oracle server key file.
keyPassword: "123%&%^$"  # Password for the key file

#Set the WS-RPC server listening interface and port of the connected Autonity Client node.
autonityWSUrl: "ws://127.0.0.1:8546"

#Set the directory of the data plugins.
pluginDir: "./plugins"  # Directory for plugins

#Set the profiling report directory, where some runtime state will be saved at.
profileDir: "."  # Profile directory

#Set the confidence strategy, available strategies are: 0: linear, 1: fixed.
confidenceStrategy: 0  # 0: linear, 1: fixed

#Set the plugin configs.
# The forex data plugins are used to fetch realtime rate of currency pairs:
# EUR-USD, JPY-USD, GBP-USD, AUD-USD, CAD-USD and SEK-USD from commercial data providers. There are 4 implemented forex
# data plugins, each of them requires the end user to apply for their own service key from the selected data provider.
# The selection of which forex data plugin(s) to use is for the end user to decide. The user can use any one of them,
# or he/she can use multiple forex data plugins in the setup. We recommend using a highly qualified data service from
# your data vendor. Avoid free or developer plans, as outdated data points may be outliers resulting in outlier penalties on your validator node's stake.
#
# The crypto data plugins are used to fetch market prices for the crypto currency pairs: ATN-USDC, NTN-USDC, NTN-ATN and
# USDC-USD. USDC liquidity is bridged to the Autonity public testnet from the Polygon Amoy testnet via a bridge service.
# Out-the-box plugins for collecting ATN-USDC and NTN-USDC market data are available for UniSwap V2 and AirSwap protocols.
# NTN-ATN market price is derived from that market data, and USDC pricing is converted to USD. ATN-NTN, ATN-USD, and
# NTN-USD prices are then submitted on-chain. To retrieve ATN and NTN prices, put the `crypto_uniswap` plugin in your plugin directory.
# Oracle server can then discover and load them. Configuring the `crypto_uniswap` plugin does not require an API key,
# it is an open and free data source of a standard EVM RPC websocket service endpoint. The end user can connect to specific
# EVM RPC endpoint base on the blockchain which hosts the uniswap contract.

# USDC-USD prices are required by the protocol to convert the ATN-USDC and NTN-USDC to ATN-USD and NTN-USD. This enables
# the reporting of ATN and NTN prices in USD to the ASM. Three plugins are implemented to source the USDC-USD datapoint
# from open and free data sources: coinbase, coingecko, and kraken. To prevent single data source failure, putting all
# 3 plugins of CEX into your plugin directory is recommended. Oracle server can then discover and load them.
# You don't need to configure the CEX plugins (crypto_coinbase, crypto_coingecko, crypto_kraken) in your oracle server
# plugin configuration file.

# For the forex data plugin default configuration is set, so the end user just needs to configure required settings,
# namely `name` and `key`. The configuration settings of a plugin are:
#

# // PluginConfig carry the configuration of plugins.
#  type PluginConfig struct {
#  Name               string `json:"name" yaml:"name"`                         // the name of the plugin binary.
#  Key                string `json:"key" yaml:"key"`                           // the API key granted by your data provider to access their data API.
#  Scheme             string `json:"scheme" yaml:"scheme"`                     // the data service scheme, http, https, ws or wss.
#  Endpoint           string `json:"endpoint" yaml:"endpoint"`                 // the data service endpoint url of the data provider.
#  Timeout            int    `json:"timeout" yaml:"timeout"`                   // the timeout period in seconds that an API request is lasting for.
#  DataUpdateInterval int    `json:"refresh" yaml:"refresh"`                   // the interval in seconds to fetch data from data provider due to rate limit.
#  NTNTokenAddress    string `json:"ntnTokenAddress" yaml:"ntnTokenAddress"`   // The NTN erc20 token address on the target blockchain.
#  ATNTokenAddress    string `json:"atnTokenAddress" yaml:"atnTokenAddress"`   // The Wrapped ATN erc20 token address on the target blockchain.
#  USDCTokenAddress   string `json:"usdcTokenAddress" yaml:"usdcTokenAddress"` // USDCx erc20 token address on the target blockchain.
#  SwapAddress        string `json:"swapAddress" yaml:"swapAddress"`           // UniSwap factory contract address or AirSwap SwapERC20 contract address on the target blockchain.
#  Disabled           bool   `json:"disabled" yaml:"disabled"`                 // The flag to disable/enable a plugin.
#}

# Un-comment below lines to enable your forex data plugin's configuration on demand.
# IMPORTANT: do not use free or developer service plan from your data vendor!
# Your production configurations start from below:
#pluginConfigs:
#  - name: forex_yahoofinance                # required, it is the plugin file name in the plugin directory.
#    key: Snp9kNMKrs8Ti ...... dIj9Y2xbPzR   # required, visit https://financeapi.net/pricing to get your key, IMPORTANT: do not use free or developer service plan.

#  - name: forex_wise                        # required, it is the plugin file name in the plugin directory.
#    key: 1234                               # required, visit https://www.wise.com to get your key, IMPORTANT: do not use free or developer service plan.
#    refresh: 300                            # optional, buffered data within 300s, recommended for API rate limited data source.

#  - name: forex_currencyfreaks              # required, it is the plugin file name in the plugin directory.
#    key: 175aab9e47e54790bf6d502c48407c10   # required, visit https://currencyfreaks.com to get your key, IMPORTANT: do not use free or developer service plan.
#    refresh: 300                           # optional, buffered data within 300s, recommended for API rate limited data source.

#  - name: forex_openexchange                # required, it is the plugin file name in the plugin directory.
#    key: 1be02ca33c4843ee968c4cedd2686f01   # required, visit https://openexchangerates.org to get your key, IMPORTANT: do not use free or developer service plan.
#    refresh: 300                           # optional, buffered data within 300s, recommended for API rate limited data source.

#  - name: forex_currencylayer               # required, it is the plugin file name in the plugin directory.
#    key: 105af082ac7f7d150c87303d4e2f049e   # required, visit https://currencylayer.com  to get your key, IMPORTANT: do not use free or developer service plan.
#    refresh: 300                           # optional, buffered data within 300s, recommended for API rate limited data source.

#  - name: forex_exchangerate                # required, it is the plugin file name in the plugin directory.
#    key: 111f04e4775bb86c20296530           # required, visit https://www.exchangerate-api.com to get your key, IMPORTANT: do not use free or developer service plan.
#    refresh: 300                           # optional, buffered data within 300s, recommended for API rate limited data source.

# Once the Autonity Genesis Foundation creates the AMM marketplace for ATN-USDC pair on Autonity blockchain, the
# foundation will announce to un-comment below lines to config the uniswap plugin to source the price of ATN-USDC.
# IMPORTANT: Do not load and config this plugin from your oracle server until officially announced.
#  - name: crypto_uniswap                          # required, it is the plugin file name in the plugin directory.
#    scheme: "wss"                                 # "wss" or "ws" please, default value is "wss" for uniswap plugin.
#    endpoint: "replace with your host:port/path"  # Change it with your validator node's web socket RPC endpoint, as the default one isn't for public usage.
#    atnTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2" # Wrapped ATN ERC20 contract address on the target blockchain.
#    usdcTokenAddress:   "0xB855D5e83363A4494e09f0Bb3152A70d3f161940" # Bridged USDC (Mainnet) / USDCx (Bakerloo) ERC20 contract address on the target blockchain.
#    swapAddress:        "0x218F76e357594C82Cc29A88B90dd67b180827c88" # UniSwap factory contract address on the target blockchain.

#Enable the metric collection for oracle server, supported TS-DB engines are influxDB v1 and v2.
#metricConfigs:
#  influxDBEndpoint: "http://localhost:8086"
#  influxDBTags: "host=localhost"
#  enableInfluxDB: false
#  influxDBDatabase: "autonity"
#  influxDBUsername: "test"
#  influxDBPassword: "test"
#  enableInfluxDBV2: false
#  influxDBToken: "test"
#  influxDBBucket: "autonity"
#  influxDBOrganization: "autonity"

```
## Data Source Strategy
When choosing a data vendor in the data API market, there are several factors to consider:

- Accuracy: The vendor must provide up-to-date pricing that accurately reflects market conditions. Therefore, it is recommended to select service plan for your data sources that offer fresh prices at least every 10 minutes. Relying on daily or hourly updates could expose your oracle client to financial penalties.

- Frequency: Given that current oracle vote round last around 10 minutes, the oracle server should provide a successful sample within this timeframe. This will help determine the rate limit level associated with your chosen data vendor.

- High Availability: For mainnet usage, it’s advisable to diversify your data sources, or at least operate two data vendors simultaneously. While a single data source may suffice, it poses a risk of omission faults if the provider encounters issues. An oracle client that experiences omission faults risks losing ATN fee rewards and NTN inflation rewards as established by the protocol.

Taking these factors into consideration, we recommend utilizing `forex_yahoofinance` with its `PRO` subscription plan for both your testnet and mainnet configurations. Additionally, it's advisable to choose a backup data provider from the available forex plugins to establish a dual data source system. Having more data vendors will enhance your availability, but it may also increase costs.

## CLI Flags
Print the version of the oracle server:
```
$./autoracle version
v0.2.5
```
Run the server:
```shell
$./autoracle ./oracle_config.yml
```

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
Prepare the plugin binaries, and save them into the `plugins` directory.
```shell
$./autoracle ./oracle_config.yml
```

#### example of profile data directory, if monitor service triggered a profile dump

```
── profiles
 └── 2024-11-19
     ├── cpu.profile_1
     ├── goroutines.txt_1
     ├── mem.profile_1
     └── trace.out_1
```

### Runtime plugin management
#### Adding new plugins
To add a new data source, just put the new plugin into the service's `plugins` directory. The oracle service auto discovers and manages it. There are no other operations required from the operator.
#### Replace running plugins
To replace running plugins with new ones, just replace the binary in the `plugins` directory. The oracle service auto discovers it by checking the modification time of the binary and does the plugin replacement itself. There are no other operations required from the operator.
#### Remove running plugins
One can remove the plugin binary from the plugin directory to remove a plugin from the server during runtime, it will also stop and unload the plugin from the oracle server.
#### Disable / Enable a plugin
A disabled plugin will be unloaded from the oracle server, one can enable it again once get the plugin and its configuration ready, then the oracle server will load and start it.

### Metrics to be collected.
#### Process Metrics
```golang
    cpuSysLoad            = GetOrRegisterGauge("system/cpu/sysload", DefaultRegistry)
    cpuSysWait            = GetOrRegisterGauge("system/cpu/syswait", DefaultRegistry)
    cpuProcLoad           = GetOrRegisterGauge("system/cpu/procload", DefaultRegistry)
    cpuThreads            = GetOrRegisterGauge("system/cpu/threads", DefaultRegistry)
    cpuGoroutines         = GetOrRegisterGauge("system/cpu/goroutines", DefaultRegistry)
    cpuSchedLatency       = getOrRegisterRuntimeHistogram("system/cpu/schedlatency", secondsToNs, nil)
    memPauses             = getOrRegisterRuntimeHistogram("system/memory/pauses", secondsToNs, nil)
    memAllocs             = GetOrRegisterMeter("system/memory/allocs", DefaultRegistry)
    memFrees              = GetOrRegisterMeter("system/memory/frees", DefaultRegistry)
    memTotal              = GetOrRegisterGauge("system/memory/held", DefaultRegistry)
    heapUsed              = GetOrRegisterGauge("system/memory/used", DefaultRegistry)
    heapObjects           = GetOrRegisterGauge("system/memory/objects", DefaultRegistry)
    diskReads             = GetOrRegisterMeter("system/disk/readcount", DefaultRegistry)
    diskReadBytes         = GetOrRegisterMeter("system/disk/readdata", DefaultRegistry)
    diskReadBytesCounter  = GetOrRegisterCounter("system/disk/readbytes", DefaultRegistry)
    diskWrites            = GetOrRegisterMeter("system/disk/writecount", DefaultRegistry)
    diskWriteBytes        = GetOrRegisterMeter("system/disk/writedata", DefaultRegistry)
    diskWriteBytesCounter = GetOrRegisterCounter("system/disk/writebytes", DefaultRegistry)
```
#### User-Plane Metrics
oracle-server metrics:
```golang
    metrics.GetOrRegisterGauge("oracle/plugins", nil) // counts the num of plugins in use.
    metrics.GetOrRegisterGauge("oracle/round", nil)   // track the current round ID.
    metrics.GetOrRegisterGauge("oracle/balance", nil) // track the current voter's account balance in ATN with 1e18 precision.
    metrics.GetOrRegisterGauge("oracle/isVoter", nil) // track if current client is a voter or not.
    metrics.GetOrRegisterCounter("oracle/l1/errs", nil) // track the num of L1 connectivity error encountered.

    // Penalize Event metrics.
    metrics.GetOrRegisterGauge("oracle/outlier_distance_percentage", nil) // track the outlier distance in percentage against the median of the round price.
    metrics.GetOrRegisterCounter("oracle/outlied_no_slashing", nil)       // track the num of outlier event which is not slashed by the protocol offensed by the server, eg.. the outlier data point is under slashing treshold of median.
    metrics.GetOrRegisterCounter("oracle/outlied_with_slashing", nil)     // track the num of outlier evwnt which is slashed by the protocol offensed by the server, eg.. the outlier data point is over slashing threshold of median.
    metrics.GetOrRegisterGaugeFloat64("oracle/slashed_ntn_total", nil)    // track the total slashed NTN stake of this server.
```
plugin metrics:
All the data points collected from the plugin are tracked in metrics with such id pattern: `oracle/plugin_name/symbol/price`:
```golang
    func (pw *PluginWrapper) updateMetrics(prices []types.Price) {
        for _, p := range prices {
            m, ok := pw.priceMetrics[p.Symbol]
            if !ok {
                name := strings.Join([]string{"oracle", pw.Name(), p.Symbol, "price"}, "/")
                gauge := metrics.GetOrRegisterGaugeFloat64(name, nil)
                gauge.Update(p.Price.InexactFloat64())
                pw.priceMetrics[p.Symbol] = gauge
                continue
            }
            m.Update(p.Price.InexactFloat64())
        }
    }
```
## Development
### Plugin Development
To add a new plugin, please visit the README.md under plugins directory, before releasing a 3rd party plugin, we will
have to check below criteria:
- **Document**    
Please add a README.md into your plugin's code directory, to describe the data source of your plugin, the data quality of it,
how can people subscribe a service key from the data provider if there is a service key required.
- **Implementation and Testing**    
Reuse the plugin framework and the standard interface of the plugin as much as possible, add test for it, help to keep the code be simple.
- **Help the user**     
In the plugin's README.md and the oracle server configuration file, add comments to guide people on how to config your
plugin, for example, the data providers' official site, how to subscribe the service key from the provider, etc...

### Build for Autonity mainnet
```shell
make autoracle
```

### Build for Bakerloo test network
```shell
make autoracle-bakerloo
```

### Build for an Autonity develop network.
If you are running a develop network for development purpose, please build the oracle components with this command:
```shell
make autoracle-dev
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

