package oracleserver

import (
	contract "autonity-oracle/contract_binder/contract"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"log"
	"math/big"
	"testing"
)

func TestCommitmentHash(t *testing.T) {
	computer, err := NewCommitmentHashComputer()
	require.NoError(t, err)

	report := []contract.IOracleReport{
		{
			Price:      common.Big1,
			Confidence: 1,
		},
	}
	salt := common.Big1

	msgSender := common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7")
	hash, err := computer.CommitmentHash(report, salt, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0x08968f6f64cc0f74029fcd9b21203ba53a59600456f4ccf58aee3476dddd39f1", hash.String())
}

func TestCommitmentHash1(t *testing.T) {
	computer, err := NewCommitmentHashComputer()
	require.NoError(t, err)
	msgSender := common.HexToAddress("0xefd5eA8c1bDC577E7e3F8172f52B42dC860a16E5")

	// round 41832
	report832 := []contract.IOracleReport{
		{
			Price:      new(big.Int).SetUint64(650444155791782400),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(726413182365303000),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1162015120140743300),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1340734186040275700),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(6778420221383200),
			Confidence: 50,
		},
		{
			Price:      new(big.Int).SetUint64(104050953335644700),
			Confidence: 50,
		},
	}

	salt832 := new(big.Int).SetUint64(3470360209374659881)
	hash832, err := computer.CommitmentHash(report832, salt832, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0x5d9d29bd0eff530a057b67dfbf65bc82438a2e7e9c3d16fd26025ab31057c243", hash832.String())

	// round41833
	report833 := []contract.IOracleReport{
		{
			Price:      new(big.Int).SetUint64(650635019779304600),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(726395241820608000),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1162350085897671300),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1341268276456852100),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(6780925257251400),
			Confidence: 50,
		},
		{
			Price:      new(big.Int).SetUint64(104085599164067700),
			Confidence: 50,
		},
	}
	salt833 := new(big.Int).SetUint64(6195178827032824501)
	hash833, err := computer.CommitmentHash(report833, salt833, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0x952eec25c193977f8d895f9d3e76a74367610672ef958c0d59b87b116e88bd65", hash833.String())

	//round 41834
	report834 := []contract.IOracleReport{
		{
			Price:      new(big.Int).SetUint64(650635019779304600),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(726395241820608000),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1162350085897671300),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1341268276456852100),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(6780925257251400),
			Confidence: 50,
		},
		{
			Price:      new(big.Int).SetUint64(104085599164067700),
			Confidence: 50,
		},
	}
	salt834 := new(big.Int).SetUint64(5663671487203855716)
	hash834, err := computer.CommitmentHash(report834, salt834, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0xda6c43a32ccfa0b9fcf845629a9a69f32919ca4cf1f2e5d19052fe642b8eb847", hash834.String())

	// round 41835
	report835 := []contract.IOracleReport{
		{
			Price:      new(big.Int).SetUint64(650635019779304600),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(726395241820608000),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1162350085897671300),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1341268276456852100),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(6780925257251400),
			Confidence: 50,
		},
		{
			Price:      new(big.Int).SetUint64(104085599164067700),
			Confidence: 50,
		},
	}
	salt835 := new(big.Int).SetUint64(976504718098674916)
	hash835, err := computer.CommitmentHash(report835, salt835, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0x31d573d9deb43ca0f4ea9dacb09cbf0be82fb6b78df73b132293772bbc285031", hash835.String())

	// round 41836
	report836 := []contract.IOracleReport{
		{
			Price:      new(big.Int).SetUint64(650635019779304600),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(726395241820608000),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1162350085897671300),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(1341268276456852100),
			Confidence: 50,
		}, {
			Price:      new(big.Int).SetUint64(6780925257251400),
			Confidence: 50,
		},
		{
			Price:      new(big.Int).SetUint64(104085599164067700),
			Confidence: 50,
		},
	}
	salt836 := new(big.Int).SetUint64(709199061146321448)
	hash836, err := computer.CommitmentHash(report836, salt836, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0x9d96665af26ce6833e490d8df1f96a2e82d70ef8dc715f065bc0e6e6ee2063e5", hash836.String())

	url := "ws://34.147.156.5:8546"
	client, err := ethclient.Dial(url)
	require.NoError(t, err)
	defer client.Close()

	height, err := client.BlockNumber(context.Background())
	require.NoError(t, err)
	log.Printf("latest block height: %d", height)

	receipt, err := client.TransactionReceipt(context.Background(), hash833)
	require.NoError(t, err)
	receipt.Logs = receipt.Logs[:]
	require.Equal(t, 0, len(receipt.Logs))
}
