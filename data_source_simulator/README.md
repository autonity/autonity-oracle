# Autonity Oracle Data Source Simulator
This component simulate data points for symbols and provide the data via HTTP rpc endpoint in the API spec of Binance.
There are two modes of simulation, default one is a random data generator where it takes a reference datapoint from config
for each symbol, then generate random data points under the configurable data distribution range around the reference point,
user can tune the reference point and distribution rate range on-demand during runtime. Another simulation mode is a simple
playbook re-player which just read the datapoint from a .csv file and keep those data point refreshing with an interval.

## Configuration
All the configuration have default values, in case of configuring the simulator, there are 3 system environment variables:
| **Env Variable**        | **Required?** | **Meaning**                                                  | **Default Value**           | **Valid Options**                |
|-------------------------|---------------|--------------------------------------------------------------|-----------------------------|----------------------------------|
| `SIM_HTTP_PORT`         | No            | The port that the simulator HTTP rpc endpoint bind to        | `50991`                     | any free port number on the host |
| `SIM_PLAYBOOK_FILE`     | No            | The data point playbook that simulator replay with           | ""                          | a .csv file with symbols at header line and datapoint at the other lines|
| `SIM_SYMBOL_CONFIG`     | No            | The string with items of patter: SYMBOL:StartingDataPoint:DataDistributionRateRange  | "NTNUSD:1.0:0.01\|NTNAUD:1.408:0.01\|NTNCAD:1.3333:0.01\|NTNEUR:0.9767:0.01\|NTNGBP:0.813:0.01\|NTNJPY:128.205:0.01\|NTNSEK:10.309:0.01"                 | similar string in such pattern |

Or, there are 5 CLI flags as well with the same feature as the system enviroment variables:

    $ ./simulator --help
    Usage of ./simulator:
    -sim_http_port=50991: The HTTP rpc port to be bind for binance_simulator simulator
    -sim_playbook_file="": The .csv file which contains datapoint for symbols.
    -sim_symbol_config="NTNUSD:7.0:0.01|NTNAUD:9.856:0.01|NTNCAD:9.333:0.01|NTNEUR:6.8369:0.01|NTNGBP:5.691:0.01|NTNJPY:128.205:0.01|NTNSEK:72.163:0.01": The list of data items with the pattern of SYMBOL:StartingDataPoint:DataDistributionRateRange with each separated by a "|"

## Deployment
Prepare the configurations via system environment variables. Create a service registration file under your service discovery DIR of the system daemon, for example "/etc/systemd/system/" in Ubuntu Linux.
Here I create a service registration file called "/etc/systemd/system/data_simulator.service" with content:
```
[Unit]
Description=Clearmatics Autonity Oracle Data Simulator
After=syslog.target network.target
[Service]
Type=simple
ExecStart=/home/jason/src/autonity-oracle/data_source_simulator/build/bin/simulator
KillMode=process
KillSignal=SIGINT
TimeoutStopSec=5
Restart=on-failure
RestartSec=5
[Install]
Alias=simulator.service
WantedBy=multi-user.target
```
### Start the service

    sudo systemctl start simulator.service

### Stop the service

    sudo systemctl stop autoracle.service

## API specifications
### Query symbol prices
Same API spec as Binance API spec for the handler of "/api/v3/ticker/price", the query string contains parameter symbols
and its value which is a JSON list of string of symbol name.

    curl -X 'GET' 'http://127.0.0.1:50991/api/v3/ticker/price?symbols=%5B%22NTNUSD%22%2C%22NTNAUD%22%2C%22NTNCAD%22%2C%22NTNEUR%22%2C%22NTNGBP%22%2C%22NTNJPY%22%2C%22NTNSEK%22%5D' -H 'accept: application/json'

### Tune the simulation
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
#### Move to new data reference point by symbols
This method move to new data reference point by symbols, thus the simulator can generate data from new reference data point.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:50991 --data '{"id":1, "method":"move_to", "params": [{"symbol": "NTNUSD", "value": 99.99},{"symbol":"NTNAUD", "value": 9.9}]}'

#### Move data reference point by percentage
This method move the data reference point of symbols by percentage, the percentage could be negative that drops the data reference point while a positive one increase
the data reference point by certain percentage, thus the simulator can generate data from new reference data point.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:50991 --data '{"id":1, "method":"move_by", "params": [{"symbol": "NTNUSD", "value": 0.01},{"symbol":"NTNAUD", "value": -0.02}]}'

#### Set data distribution rate range
This method set new data distribution rate range for symbols' data generator, thus the simulator can gen data with the data distribution range.

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:50991 --data '{"id":1, "method":"set_distribution_rate", "params": [{"symbol": "NTNUSD", "value": 0.01},{"symbol":"NTNAUD", "value": 0.02}]}'

#### Simulate new symbols
This method simulate data points for new symbols, the value field specifies the data reference point for the symbols, while the distribution rate is resolved by default value 0.01. One can change it by calling RPC call "set_distribution_rate".

    curl -X POST -H "Content-Type: application/json" http://127.0.0.1:50991 --data '{"id":1, "method":"new_simulation", "params": [{"symbol": "NTNUSD", "value": 99.77},{"symbol":"NTNAUD", "value": 0.92}]}'

