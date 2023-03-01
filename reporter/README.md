# Data Reporting Module
To report oracle data to an on-chain oracle contract from oracle client, we need to build a reporting module that bind with oracle contract to report round data.

## Work flows in Oracle Client
![Screenshot from 2023-02-28 07-01-27](https://user-images.githubusercontent.com/54585152/221778482-2939116f-1025-4013-beac-31f934430c10.png)

## Oracle Contract
### Deployment
Introduced a new parameter, the OracleClient, for validator object in the genesis file, it describes the binding relationship between a validator and its corresponding oracle client which serves the oracle data service for it.
#### committees in the bootstrap
Options
    - Return the committee from autonity contract's constructor.
    - Compute the committee by voting power and commitee size as the contructor of autonity contract.
#### binding in the bootstrap
Get bindings from genesis config.

#### The constructor of Oracle contract
requires the binding information and the committee to initialize the oracle contract.

#### keep binding upated
    - setComitte should bring bindings of committee members.
    - updateBinding once a validator update its binding in the validator pool of autonity contract.

### Data reporting
#### The interface
A new parameter introduced, the _round which represents the round that the commitment refers to, and the _reports was the previous round data measured in the last round, _round - 1.
    
    function vote(uint256 _round, uint256 _commit, int[] memory _reports, uint256 _salt)
    
#### Bootstrap
Vote with commitment of current round without last round data.

#### Node recover from a restart.
Vote with commitment of current round without last round data.

#### Epoch Rotation
    - New validator added in the committee
        Vote with commitment of current round but without last round data.
    - The validator who keeps in the new committee
        Vote with commitment of current round and last round data.
    - The validator who is no longer in the new committee
        Vote without commitment of current round but with last round data.

All of the above data reporting scenario is valid and normal reports. Any missing of report from this set is accountable for penalty protocol.
