# Autonity Oracle Component 

## Assumptions 

This project assumes the following:

* Go 1.19.3 
* Linux / MacOS operating system

## Product introduction
This component works as the bridge that brings data points from different data provider and unifies the data into the standard format that can be pushed to Autonity L1 network. To support the runtime adaptations with different data providers, the adapters are implemented in plugins mechanism thus it maintains high availability of the autonity oracle service. 
As the component starts ticker jobs that fetch data points from providers on every 10s timely, it also scans the plugin directory on every 2s to launch new plugins or to replace runtime plugins with newly modified one to adapt with data provider. The data aggregation at this level is base on taking the median value out from the data set. By providing unified data points and pushing the data samples on a round base intervals to L1 oracle contract, the autonity layer1 network can get the requried data for its stabilisation protocol.

![Screenshot from 2023-02-28 06-55-34](https://user-images.githubusercontent.com/54585152/221777340-400bedc7-6a55-4055-8b18-4f6158e168c6.png)

## Configuration 
Values that can be configured by using environment variables:    

| **Env Variable**        | **Required?** | **Meaning**                                                  | **Default Value**           | **Valid Options**                |
|-------------------------|---------------|--------------------------------------------------------------|-----------------------------|----------------------------------|
| `ORACLE_HTTP_PORT`      | No            | The port that the HTTP service endpoint bind to              | `30311`                     | any free port number on the host |
| `ORACLE_CRYPTO_SYMBOLS` | No            | The symbols that the oracle component collect data point for | "ETHUSDC,ETHUSDT,ETHBTC"    | symbols seperated by ','         |
| `ORACLE_PLUGIN_DIR`     | No            | The directory that stores the plugins                        | "./plugins"                 | any directory that saves plugins |
| `GIN_MODE`              | No            | The mode running by the HTTP service                         | "debug"                     | release or debug                 |

or by using console flags:

    $./autoracle -help
    Usage of ./autoracle:
    -oracle_crypto_symbols="ETHUSDC,ETHUSDT,ETHBTC": The symbols string separated by comma
    -oracle_http_port=30311: The HTTP service port to be bind for oracle service
    -oracle_plugin_dir="./plugins": The DIR where the adapter plugins are stored

example to run the autonity oracle service with console flags:
    
    $./autoracle -oracle_crypto_symbols="ETHUSDC,ETHUSDT,ETHBTC" -oracle_http_port=30311 -oracle_plugin_dir="./plugins"

## Developing

To run e2e test use

    make e2e-test

To run all tests use
    
    make test

To generate code coverage reports run

    make test-coverage

To lint code run

    make lint

To build the data source simulator run

    make simulator

To build the project run

    make autoracle

The built binaries are presented at: ./build/bin under which there is a plugins directory saves the built plugins as well.

## Deployment
### Start up the service from shell console
Prepare the plugin binaries, and save them into the plugin directory, then start the service:
Set the system environment variables and run the binary:

    $export ORACLE_HTTP_PORT=63306
    $export ORACLE_CRYPTO_SYMBOLS="ETHUSDC,ETHUSDT,ETHBTC"
    $export ORACLE_PLUGIN_DIR="./plugins"
    $.~/src/autonity-oracle/build/bin/autoracle    

or configure by using console flags and run the binary:

    $.~/src/autonity-oracle/build/bin/autoracle -oracle_crypto_symbols="ETHUSDC,ETHUSDT,ETHBTC" -oracle_http_port=63306 -oracle_plugin_dir="./plugins"

### An elegant way base on linux system daemon
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
For new adaptations with newly added plugins, just put the new plugins into the service's plugin directory, the service auto discovery it and manage it. There is no other operations are required from operator.
#### Replace running plugins
To replace running plugins with new ones, just replace the binary under the plugin directory, the service auto discovery it by checking the modification time of the binary and do the plugin replacement itself, there is no other operations are required from operator.

## API specifications
The HTTP request message and response message are defined in json object JSONRPCMessage, it is carried by the HTTP body in both the request or response message, all the APIs are access with POST method by specifying the method and the corresponding method's params in params field, and the ID help the client to identify the requests and response pairing.
```go
    type JSONRPCMessage struct {
        ID     json.RawMessage `json:"id,omitempty"`
        Method string          `json:"method,omitempty"`
        Params json.RawMessage `json:"params,omitempty"`
        Result json.RawMessage `json:"result,omitempty"`
        Error  string          `json:"error,omitempty"`
    }
```

### get_version
This method return the oracle service version.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:63306 --data '{"id":1, "method":"get_version", "params": []}'

```json
{"id":1,"result":{"Version":"0.0.1"}}
```    
### get_prices
This method returns all the symbols corresponding prices, and also the current symbols that is used by the oracle service.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:63306 --data '{"id":1, "method":"get_prices", "params": []}'

```json
{"id":1,"result":{"Prices":{"NTNBTC":{"Timestamp":1672836542504,"Symbol":"NTNBTC","Price":"11.11"},"NTNETH":{"Timestamp":1672836542504,"Symbol":"NTNETH","Price":"11.11"},"NTNRMB":{"Timestamp":1672836542504,"Symbol":"NTNRMB","Price":"11.11"}},"Symbols":["NTNBTC","NTNETH","NTNRMB"]}}
```
### update_symbols
This method update the symbols of current oracle service, and returned the updated symbols once update is finished.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:63306 --data '{"id":1, "method":"update_symbols", "params": ["NTNUSDC,NTNUSDT,NTNDAI"]}'

```json
{"id":1,"result":["NTNUSDC,NTNUSDT,NTNDAI"]}
```

### list_plugins
This method list all the running plugins on the oracle service.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:30311 --data '{"id":1, "method":"list_plugins", "params": []}'

```json
{"id":1,"result":{"binance":{"Version":"v0.0.1","Name":"binance","StartAt":"2023-01-12T11:43:10.32010817Z"},"fakeplugin":{"Version":"v0.0.1","Name":"fakeplugin","StartAt":"2023-01-12T11:43:10.325786993Z"}}}
```
