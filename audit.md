# Autonity Oracle Server Audit Scope
**Liveness**    
As a side component of the Autonity validator, oracle-server should run with high availability. 
1. The connectivity with Autonity Validator can self heal by itself.
   `./oracle_server/oracle_server.go:checkHealth()`
2. There are no deadlocks during the data collection and voting period.
   `./oracler_server`, `./plugin_wrapper`, `./plugins/crypto_*`, `./plugins/forex_*`
3. The voter's certification are always synced.
 `./oracle_server/oracle_server.go:handleRoundVote()`, `./oracle_server/oracle_server.go:isVoter()`, `./oracle_server/oracle_server.go:syncStates()`
4. The plugins should run with high availability as well.
   `./oracle_server/oracle_server.go:PluginRuntimeDiscovery()`,`./plugin_wrapper`, `./plugins/crypto_*`, `./plugins/forex_*`

**Safety**    
1. On-chain state sync and round coordination.
   `./oracle_server/oracle_server.go:handleRoundVote()`, `./oracle_server/oracle_server.go:isVoter()`, `./oracle_server/oracle_server.go:syncStates()`
3. To prevent from free riding, the commit-reveal mechanism is correctly implemented.
   `./oracle_server/`, `./oracle_server/commitment_hash_computer.go`
4. The off-chain data point aggregation.
   `./oracle_server/oracle_server.go:aggregateProtocolSymbolPrices()`
6. Data sampling with rate limit and delays.
   `./oracle_server/oracle_server.go:handlePreSampling()`, `./plugins/common/common.go:FetchPrices()`
   Check if delays and rate limit of the data source can introduce high risk of penalty. Especially for the datapoint sampling of AMM, we want to reduce the risk AMSP.
8. Self protection.
   Under implemenation.


