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
The HTTP body contains a json object: 
### get_version

### get_prices

### update_symbols
