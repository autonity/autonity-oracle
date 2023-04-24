# Autonity Oracle Network 

## Assumptions 

This project assumes the following:

* Go 1.19.3 
* Linux operating system

## Overview
The Autonity oracle network provides exchange rate data for currency pairs sourced from different data providers and unifies these data points into a
standard format that can be pushed to the oracle contract deployed on the Autonity L1 network. The oracle contract aggregates th0se external data points to provide reference exchange rate data agreed by L1 consensus as an L1 feature. DApps deployed on the Autonity L1 network can consume these data points via oracle contract interfaces.  

## The oracle client operator
The oracle client is operated and maintained by the autonity validator node operator. Validator node operators are required to provide an ownership proof for the oracle client when submitting a validator registration transaction to the Autonity Contract. Thus the oracle client and the validator node  client participate in the oracle protocol and the autonity protocol to provide the oracle data service and the L1 block validation service.

## Data adaptors - plugin architecture
The Autonity oracle client provides a unified interface between itself and plugins that adapt data from different data providers, for example Binance and Coingecko, etc. Any party can build a plugin implementing this unified interface and so provide an adaptor for any data source on demand. The oracle client will scan and load plugins from the plugin directory during runtime. Detection of new or changed plugins is dynamic; no shutdown of the oracle client is required to detect and apply the change.

## Coordination of data sampling
To coordinate data sampling in the oracle network, the L1 oracle contract issues a round event on every vote period (60 blocks). The round event carries a tuple `(RoundID, SampleTS, Height, VotePeriod)`, which tell the oracle clients that on round with ID `RoundID`, a data sample with timestamp `SampleTS` is required for the data submission. The `Height` stands for the start height of the new round, while the `VotePeriod` stands for the round length of the new round. Thus the oracle client can estimate and manage data pre-samplings for the new round and then pick up the nearest sample refering to the required `SampleTS`. 

![Screenshot from 2023-04-21 04-19-10](https://user-images.githubusercontent.com/54585152/233533092-29b65a39-eb87-496f-9a1e-0741bc7fbd45.png)

## Configuration 
Values that can be configured by using environment variables:    

| **Env Variable**           | **Required?** | **Meaning**                                                                                 | **Default Value**                   | **Valid Options**                                       |
|----------------------------|---------------|---------------------------------------------------------------------------------------------|-------------------------------------|---------------------------------------------------------|
| `ORACLE_CRYPTO_SYMBOLS`    | No            | The symbols that the oracle component collects data points for                                | "NTNUSD,NTNAUD,NTNCAD,NTNEUR,NTNGBP,NTNJPY,NTNSEK"            | symbols seperated by ','                                |
| `ORACLE_PLUGIN_DIR`        | No            | The directory that stores the plugins                                                       | "./build/bin/plugins"                         | any directory that saves plugins                        |
| `ORACLE_KEY_FILE`          | Yes           | The encrypted key file path that contains the private key of the oracle client.             | "./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" | any key file that saves the private key                 |
| `ORACLE_KEY_PASSWORD`      | Yes           | The password of the key file that contains the private key of the oracle client.            | "123"                               | any password that encrypted the private key             |
| `ORACLE_AUTONITY_WS_URL`   | Yes           | The web socket RPC URL of your Autonity L1 Node that the oracle client communicates with.   | "ws://127.0.0.1:8000"               | the web socket rpc endpoint url of the Autonity client. |

or by using console flags:

    $./autoracle -help
    Usage of ./autoracle:
    -oracle_autonity_ws_url="ws://127.0.0.1:8000": The websocket URL of autonity client
    -oracle_crypto_symbols="ETHUSDC,ETHUSDT,ETHBTC": The symbols string separated by comma
    -oracle_key_file="a path to your key file": The file that save the private key of the oracle client
    -oracle_key_password="key-password": The password to decode your oracle account's key file
    -oracle_plugin_dir="./plugins": The DIR where the adapter plugins are stored


example to run the autonity oracle service with console flags:
    
    $./autoracle -oracle_crypto_symbols="ETHUSDC,ETHUSDT,ETHBTC" -oracle_plugin_dir="./plugins" -oracle_key_file="../../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" -oracle_key_password="123" -oracle_autonity_ws_url="ws://127.0.0.1:8000"

## Developing

To build the project run

    make autoracle

To build the data source simulator run

    make simulator

To run e2e test use

    make e2e-test

To run all tests use
    
    make test

To lint code run

    make lint

To generate mocks for unit test
    
    make mock



Built binaries are presented at: `./build/bin` under which there is a `plugins` directory for the built plugins as well.

## Deployment
### Oracle Client Private Key generation
Download the Autonity client to generate the private key from console, and set the password to encode the key file, the
key file path will display, and remember the password that encrypted the key file.

    $./autonity --datadir ./keys/ account new
    Your new account is locked with a password. Please give a password. Do not forget this password.
    Password:xxxxxx
    Repeat password:xxxxxx

    Your new key was generated

    Public address of the key:   0x7C785Fe9404574AaC7daf2FF30637546493900d1
    Path of the secret key file: key-data/keystore/UTC--2023-02-28T11-40-15.383709761Z--7c785fe9404574aac7daf2ff30637546493900d1

    - You can share your public address with anyone. Others need it to interact with you.
    - You must NEVER share the secret key with anyone! The key controls access to your funds!
    - You must BACKUP your key file! Without the key, it's impossible to access account funds!
    - You must REMEMBER your password! Without the password, it's impossible to decrypt the key!

### Start up the service from shell console
Prepare the plugin binaries, and save them into the `plugins` directory, then start the service:
Set the system environment variables and run the binary:

    $export ORACLE_CRYPTO_SYMBOLS="ETHUSDC,ETHUSDT,ETHBTC"
    $export ORACLE_PLUGIN_DIR="./plugins"
    $export ORACLE_KEY_FILE="./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
    $export ORACLE_KEY_PASSWORD="your passord to the key file"
    $export ORACLE_AUTONITY_WS_URL="ws://127.0.0.1:8645"
    $.~/src/autonity-oracle/build/bin/autoracle

or configure by using console flags and run the binary:

    $./autoracle -oracle_crypto_symbols="ETHUSDC,ETHUSDT,ETHBTC" -oracle_plugin_dir="./plugins" -oracle_key_file="../../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe" -oracle_key_password="123" -oracle_autonity_ws_url="ws://127.0.0.1:8645"

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
ExecStart=/home/jason/src/autonity-oracle/build/bin/autoracle -oracle_plugin_dir="/home/jason/src/autonity-oracle/build/bin/plugins"
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

    sudo systemctl start autoracle.service

#### Stop the service

    sudo systemctl stop autoracle.service

```
● autoracle.service - Clearmatics Autonity Oracle Server
     Loaded: loaded (/etc/systemd/system/autoracle.service; disabled; vendor preset: enabled)
     Active: inactive (dead)

Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th systemd[1]: Stopping Clearmatics Autonity Oracle Server...
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.233Z [INFO]  *oracleserver.OracleServer: the jobTicker jobs of oracle service is stopped
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.233Z [DEBUG] binance.binance: 2023/01/19 03:03:45 [DEBUG] plugin: plugin server: accept unix /tmp/plugin3024381010: use of closed network connection
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.235Z [INFO]  binance: plugin process exited: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance pid=14577
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.235Z [DEBUG] binance: plugin exited
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.236Z [DEBUG] fakeplugin.fakeplugin: 2023/01/19 03:03:45 [DEBUG] plugin: plugin server: accept unix /tmp/plugin2424636505: use of closed network connection
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.237Z [INFO]  fakeplugin: plugin process exited: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin pid=14586
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T03:03:45.237Z [DEBUG] fakeplugin: plugin exited
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th systemd[1]: autoracle.service: Succeeded.
Jan 19 03:03:45 jason-ThinkPad-X1-Carbon-7th systemd[1]: Stopped Clearmatics Autonity Oracle Server.

```

#### Check the runtime status

    sudo systemctl status autoracle.service

```
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
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]:          Running autonity oracle service at port: 30311, with symbols: NTNUSDT,NTNUSDC,NTNBTC,NTNETH and plugin diretory: /home/jason/src/autonity-oracle/build/bin/plugins
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [WARN]  binance: plugin configured with a nil SecureConfig
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: starting plugin: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance args=[/home/jason/src/autonity-oracle/build/bin/plugins/binance]
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: plugin started: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance pid=14577
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.152Z [DEBUG] binance: waiting for RPC address: path=/home/jason/src/autonity-oracle/build/bin/plugins/binance
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.159Z [DEBUG] binance.binance: plugin address: network=unix address=/tmp/plugin3024381010 timestamp=2023-01-19T02:42:19.159Z
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.159Z [DEBUG] binance: using plugin: version=1
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [INFO]  binance: plugin initialized: binance=v0.0.1
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [WARN]  fakeplugin: plugin configured with a nil SecureConfig
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: starting plugin: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin args=[/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin]
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: plugin started: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin pid=14586
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.161Z [DEBUG] fakeplugin: waiting for RPC address: path=/home/jason/src/autonity-oracle/build/bin/plugins/fakeplugin
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.168Z [DEBUG] fakeplugin.fakeplugin: plugin address: address=/tmp/plugin2424636505 network=unix timestamp=2023-01-19T02:42:19.167Z
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.168Z [DEBUG] fakeplugin: using plugin: version=1
Jan 19 02:42:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:19.170Z [INFO]  fakeplugin: plugin initialized: fakeplugin=v0.0.1
Jan 19 02:42:29 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:42:29.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:42:29.156Z
Jan 19 02:43:19 jason-ThinkPad-X1-Carbon-7th autoracle[14568]: 2023-01-19T02:43:19.156Z [DEBUG] fakeplugin.fakeplugin: receive request from oracle service, send data response: timestamp=2023-01-19T02:43:19.156Z
```

### Runtime plugin management
#### Adding new plugins
To add a new data source, just put the new plugin into the service's `plugins` directory. The oracle service auto discovers and manages it. There are no other operations required from the operator.
#### Replace running plugins
To replace running plugins with new ones, just replace the binary in the `plugins` directory. The oracle service auto discovers it by checking the modification time of the binary and does the plugin replacement itself. There are no other operations required from the operator.
