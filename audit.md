# Autonity Oracle Protocol Audit Scope

## On-chain oracle protocol contract
**Liveness**    
As the oracle protocol is coordinated by the block finalisation function, at each block finalization phase, we need to make sure the state transition of the oracle contract are properly process. There shouldn't be any un-expected revert from the EVM which cause the blockchain be halted. With such requirement, we need to review the contract to find out any edge cases which are not covered by current logic. It will include:
1. The vote message processing with commit-reveal machanism.
2. The on-chain datapoint aggregation.
3. The outlier detection.
4. The reward distribution.
5. The slashing penalty.
6. The voter's certification.    

**Safety**     
As the required final data point are used by Autonity's ASM, which is the economic core of the Autonity network, we want to make sure that the values aggregated by the protocol are correct, it should represent the real world's state of the data point of the corresponding currency symbols. We need to address any market data manipulation misbehaviours.

## Off-chain oracle server
**Liveness**    
As a side component of the Autonity validator, oracle-server should run with high availability. 
1. The connectivity with Autonity Validator can self heal by itself.
2. There are no deadlocks during the data collection and voting period.
3. The voter's certification are always synced.
4. The plugins should run with high availability as well.

**Safety**    
1. Only cerficated voter are allowed to vote.
2. To prevent from free riding, the commit-reveal mechanism is correctly implemented.
3. The off-chain data point aggregation.
4. Data sampling with rate limit and delays.
   Check if delays and rate limit of the data source can introduce high risk of penalty. Especially for the datapoint sampling of AMM, we want to reduce the risk AMSP.
5. Self protection.
   If an oracle node get slashed as an outlier, it should prevent from getting slashed again.


