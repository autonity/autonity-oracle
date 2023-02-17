// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 < 0.9.0;

// todo: implement the oracle contract for the protocol.

/**
*  @title Oracle Contract
*  @dev This would likely be set behind an open zeppeling proxy contract OR we
*  should keep our upgrade mechanism
*/

contract Oracle {
    // This is a special value representing unable to fetch the correct data.
    //uint256 internal constant INVALID_PRICE = uint256(-1);
    uint256 internal constant VOTING_GAS = 30_000;

    struct Vote {
        uint vote;   // index of the voted proposal
    }

    struct Prevote {
        string symbol;
        uint256 rate;
    }


    string[] public symbols;
    string[] public newSymbols;

    // oracle client may listen for these events to take actions.
    event UpdatedSymbols(string[] symbols);
    event UpdatedRound(uint256 round);
    event UpdatedCommittee(address[] committee);

    address private autonity;
    address[] private committee;
    uint256 public round;

    mapping(address => uint256) public commits;
    mapping(string => mapping(address => uint256)) public prevotes;

    mapping(address => uint256) public voted;

    struct Price {
        int256 price;
        uint timestamp;
        uint status; // Do we get back a status code if couldn't compute last
        // price in time ?
    }

    mapping(string => Price)[] internal prices;

    constructor(address[]memory _committee) {
        autonity = msg.sender;
        committee = _committee;
        // all of them are added
    }

    /**
    * @notice Vote for the current period. In order to save gas,
    * if (prevotes[i] == INVALID_PRICE){g the symbols.
    * if the validator leave consensus committee then his vote is discarded.
    * if a validator joins the consensus committee then his first vote is not
    * taken into account.
    * Only allowed to vote once per round.
    * @param _commit hash of the new prevotes
    *        _prevotes reveal of the prevotes for the previous cycle.
    */
    function vote(uint256 _commit, int[]memory _prevotes) external {
        /*
        require(voted[msg.sender] < round, "already voted");
        // revert if data is not supplied -
        //require(_prevotes.lengths == symbols.lengths, "missing data");
        // Edge scenario for first epoch to be taken care of.
        uint256 _pastCommit = commits[msg.sender];
        // Check that reveal matches past commit. TODO: include salt.
        if (_pastCommit != keccak256(abi.encodePacked(_prevotes))) {
            // Reveal is not matching commit here, do something
            msg.sender;
        }
        commits[msg.sender] = _commit;

        // Voter has to vote on all the symbols
        // uint256 MAX_INT = uint256(-1) is a special value
        for (uint256 i = 0; i < _prevotes.length; i++) {
            prevotes[symbols[i]][msg.sender] = _prevotes[i];
        }

        voted[msg.sender] = round;

        // TODO: How to ensure there would be always sufficient
        // funds to reimburse voting? Should AUT be minted here ?
        // Also: consider tx.gasPrice versus block.baseFee
        msg.sender.transfer(block.baseFee * VOTING_GAS);
        */
    }

    /**
     * @notice Called once per VotePeriod part of the state finalisation function.
     *
     */
    function finalize() public {
        emit UpdatedRound(round);
        return;
    }

    /**
     * @notice Level 2 aggregation routine. For the time being, the final price
     * is a simple average.
     * @dev This method is responsible in detecting and calling the appropriate
     * accountability functions in case
     * of missing or malicious votes.
     */
    function aggregateSymbol(uint _sindex) internal {
        /*
        string _symbol = symbols[_sindex];
        // Final aggregation doesn't depend on price.
        // Sort prices.
        int256 _total;
        int256 _count;
        for(uint i = 0; i < committee.length; i++) {
            if(voted[committee[i]] != round) {
                // TODO: Implement
                autonity.oracleVoteMissing(committee[i]);
                continue;
            }
            _total += votes[_symbols][committee[i]];
            _count += 1;

        }
        prices[round][_symbol] = Price(
            _total / _count, // average
            block.timestamp,
            1);
        */
    }

    /**
     * @notice Return latest available price data.
     * @param _symbol, the symbol from which the current price should be returned.
     */
    function latestData(string memory _symbol) public view returns (int256, uint, uint)
    {
        Price memory _p = prices[round][_symbol];
        return (_p.price, _p.timestamp, _p.status);
    }

    function getRoundData(uint256 _round, string memory _symbol) external view returns
    (int256, uint, uint){
        return (0, 0, 0);
    }

    function setSymbols(string[]memory _symbols) public {
        symbols = _symbols;
        emit UpdatedSymbols(_symbols);
    }

    function setCommittee(address[]memory _newCommittee)  public {
        committee = _newCommittee;
        emit UpdatedCommittee(_newCommittee);
    }

    function getCommittee() public view returns (address[]memory) {
        return committee;
    }

    function getRound() public view returns (uint256) {
        return round;
    }

    function getSymbols() public view returns (string[]memory) {
        return symbols;
    }
}