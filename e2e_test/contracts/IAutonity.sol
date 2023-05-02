// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 < 0.9.0;

/**
 * @dev Interface of the Autonity Contract that is used for integration test only.
 */
interface IAutonity{
    struct CommitteeMember {
        address addr;
        uint256 votingPower;
    }
    /**
     * @notice Retrieve the lists of current committee members with their corresponding voting power.
     */
    function getCommittee() external view returns (CommitteeMember[] memory);

    /**
     * @notice set the committee size, only used by system operator for integration test.
     */
    function setCommitteeSize(uint256 _size) external;
}
