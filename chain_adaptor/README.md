# Data Reporting Module
To report oracle data to an on-chain oracle contract from oracle client, we need to build the reporting module.

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


## Work flows
![Screenshot from 2023-02-10 10-30-08](https://user-images.githubusercontent.com/54585152/218069488-97345965-31e3-40ed-895b-af8c311b1f3d.png)

### Oracle contract deployment
Deploy oracle contract after the autonity contract is deployed, save the oracle contract address as a protocol parameter for discovery.

### Data reporting.
On the oracle client start up, it subscribes the chain head event, and get the oracle contract address. If it is a participant of current consensus committee,
it submits data report if it does not send any for current round, round is resolved by epoch length and round length. The reporting is base on a commit and reveal
mechanism. A similar one is explain here: https://github.com/clearmatics/autonity-protocol/discussions/187 with a commitment hash and a report to reveal the validity of report.

### Data Finalization.
At the end of round, the protocol calls the function: finalize() in the oracle contract that will resolve the price of all the symbols for the time of current round being,
thus the external contracts or the stabilisation module can target to the price of each required symbols.

### Accountability, incentives and slashing.
On each data report, the oracle contract collects the number of valid report rounds of each reporter, thus we have the live-ness of oracle client collected for each epoch,
so in the end of epoch, the reward distribution and slashing can happen by according to the collected accountability data. We need to kick off the incentive and slashing
sub protocol on this part.

### Epoch rotation.
With committee member re-shuffled, as the oracle client subscribes the chain head event, they can be notified with the role changes, those client who were not a member of
committee start to skip the report while those one in the committee set should start to report data.

## Task dependency

### Phase 1: Oracle Contract ABI Spec.
### Phase 2: Contract implementation and the interactions between contract and oracle client.
1. The implementation of contract and deploy it.
2. The interactions between contract and oracle client.
    - a chain event subscribe base on web socket.
    - a go binder of oracle contract.
    - data reporting on commit and reveal way.
### Phase integration test
### E2E test
