# Chain Adaptor
To report oracle data to an on-chain oracle contract from oracle client, we need to build the antonity L1 blockchain adaptor.

## Basic helper functions

### Chain Head Event Subscriber

#### Resolve the latest symbols required by the protocol
The oracle client need to subscribe the chain head event to coordinate the data submission.
#### Resolve the current committee
Only committee member need to report the data points.
#### Resolve the data reporting round
An oracle just need report once for each required symbol at a round.
#### Use the Chain Height to coordinate the timing of data report.
At a certain height of a round, the oracle should send the TX to submit the report.

### Oracle Contract Go Binder
An Oracle contract Go Binder should be generated to interact with the oracle contract.
Get the contract ABI from Autonity L1 rpc endpoint to do the data packing.

### A key store to manage the node key of oracle service.
The key is used to sign the TX that contains the data report collected by the oracle service.

### The oracle contract
An oracle contract collects data reports on each round, and finalize the price for symbols at the end of round, a round
is fix serial of blocks under an epoch. Those finalized prices of symbols are exposed for the other DApps include stabilisation module.

#### Contract address discovery
The Autonity client should provide the oracle contract address via an RPC API, the L1 protocol can deploy the oracle contract right after the genesis block.
#### Collecting and Finalizing price for symbols at round base.
Require a data aggregation algorithm to compute the resolved price, currently we assume a median value base algorithm is fine.

#### Accountability
A data struct to track those omission faults of data reporting. The normal ones are reward by the autonity protocol.
#### Free riding issue
Use the commit and reveal mechanism to prevent from free riding.

### Incentives
L1 protocol should reward the oracle client via native token that they get refund for data reporting cost.

### Slashing
Those omission faulty oracle client's binding validator would be slashed.