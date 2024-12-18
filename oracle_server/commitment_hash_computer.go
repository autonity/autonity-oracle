package oracleserver

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var ReportsSchema = []byte("[{\"components\":[{\"internalType\":\"uint120\",\"name\":\"price\",\"type\":\"uint120\"},{\"internalType\":\"uint8\",\"name\":\"confidence\",\"type\":\"uint8\"}],\"internalType\":\"struct Report[]\",\"name\":\"_reports\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_salt\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}]")

type CommitmentHashComputer struct {
	args abi.Arguments
}

func NewCommitmentHashComputer() (*CommitmentHashComputer, error) {
	var args abi.Arguments
	err := json.Unmarshal(ReportsSchema, &args)
	if err != nil {
		return nil, err
	}

	return &CommitmentHashComputer{args: args}, nil
}

// CommitmentHash computes the keccak256Hash of the output of the solidity instruct `abi.encode(_reports, _salt, msg.sender)`
func (c *CommitmentHashComputer) CommitmentHash(args ...interface{}) (common.Hash, error) {
	var hash common.Hash
	bytes, err := c.args.PackValues(args)
	if err != nil {
		return hash, err
	}

	hash = crypto.Keccak256Hash(bytes)
	return hash, nil
}
