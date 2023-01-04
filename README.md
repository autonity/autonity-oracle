# Autonity Oracle Component 

## Assumptions 

This project assumes the following:

* Go 1.19.3 
* Linux / MacOS operating system

## Product introduction 
This component works as the bridge that brings data points from different data provider and unifies the data into the standard format that exposed by HTTP service. Thus, the Autonity blockchain nodes can fetch the unified data points from this component for its stabilisation module.
The component starts a ticker job that is triggered every 10s timely to fetch data point from the data providers it adapted for those symbols which are interested of by the Autonity blockchain for its stabilisation module, meantime the component provides the unified price by symbols via the HTTP service.

## Configuration 
Values that can be configured using environment variables:    

| **Env Variable**        | **Required?** | **Meaning**                                                  | **Default Value**  | **Valid Options**                |
|-------------------------|---------------|--------------------------------------------------------------|--------------------|----------------------------------|
| `ORACLE_HTTP_PORT`      | Yes           | The port that the HTTP service endpoint bind to              | `30311`            | any free port number on the host |
| `ORACLE_CRYPTO_SYMBOLS` | Yes           | The symbols that the oracle component collect data point for | \"BNBBTC,BTCUSDT\" | symbols seperated by ','         |
| `GIN_MODE`              | No            | The mode running by the HTTP service                         | "debug"            | release or debug                 |

## Developing

To run e2e test use

    make e2e_test

To run uint tests use
    
    make test

To generate code coverage reports run

    make test_coverage

To lint code run

    make lint

To build the project run

    make autoracle

## Deployment

Set the system environment variables and run the binary:

    $export ORACLE_HTTP_PORT=63306
    $export ORACLE_CRYPTO_SYMBOLS="NTNBTC,NTNETH,NTNRMB"
    $.~/src/autonity-oracle/build/bin/autoracle    


## API specifications
The HTTP request message and response message are defined in json object JSONRPCMessage, it is carried by the HTTP body      
in both the request or response message, all the APIs are access with POST method by specifying the method and the     
corresponding method's params in params field, and the ID help the client to identify the requests and response pairing.
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