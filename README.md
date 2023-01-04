# Autonity Oracle Component 

## Assumptions 

Details runtime and operating system dependecies. 

This project assumes the following:

* Go 1.19.3 
* Linux / MacOS operating system

## Product introduction 
todo: write the features of this component

## Configuration 
Values that can be configured using environment variables:    

| **Env Variable**                                | **Required?** | **Meaning**                                                  | **Default Value**  | **Valid Options**                 |
|-------------------------------------------------|---------------|--------------------------------------------------------------|--------------------|-----------------------------------|
| `ORACLE_HTTP_PORT`                              | Yes           | The port that the HTTP service endpoint bind to              | `30311`            | any free port number on the host  |
| `ORACLE_CRYPTO_SYMBOLS`                         | Yes           | The symbols that the oracle component collect data point for | \"BNBBTC,BTCUSDT\" | symbols seperated by ','          |

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

Set the system environment variables and then run the binary autoracle.

todo: write some example to start the component.


