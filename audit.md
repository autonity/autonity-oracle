# Autonity Oracle Server Audit Scope
Below are the core functionality of oracle server:     
**core of oracle server**     
`./oracle_server/oracle_server.go` needs to be fully audited. The core logic is:
1. The connectivity with Autonity Validator can self heal by itself.
2. The plugins runtime management.
3. On-chain state sync and round coordination.
   The state of symbols, voters, round period, are synced with oracle contract.
4. The prevent of free riding.
   a commit-reveal mechanism was designed to prevent it.`./oracle_server/commitment_hash_computer.go`
5. The off-chain data point aggregation.
   `./oracle_server/oracle_server.go:aggregateProtocolSymbolPrices()`
6. Data sampling with rate limit and delays.
   Check if delays and rate limit of the data source can introduce high risk of penalty. Especially for the datapoint sampling of AMM, we want to reduce the risk AMSP. `./oracle_server/oracle_server.go:handlePreSampling()`, `./plugins/common/common.go:FetchPrices()`

**plugins**    
   `./oracler_server`, `./plugin_wrapper`, `./plugins/crypto_uniswap`
   The AMM plugin, for example the uniswapV2 plugin is the most important content for plugins.
   Also the plugin framework should be checked as well. `./plugins/common/`, there is a readme for the framework `./plugins/README.md`

Apart from these we would like to make sure if it's safe from exploits and there are no possibilities of crashes or deadlocks due to external/internal factors (e.g.) 
