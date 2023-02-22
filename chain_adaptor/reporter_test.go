package chain_adaptor

import (
	"autonity-oracle/config"
	oracleserver "autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDataReporter(t *testing.T) {
	priKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	ws := config.DefaultAutonityWSUrl
	validator := common.Address{}
	key := &keystore.Key{
		PrivateKey: priKey,
		Address:    crypto.PubkeyToAddress(priKey.PublicKey),
	}
	oracle := &oracleserver.OracleServer{}

	t.Run("new data report", func(t *testing.T) {
		dp := NewDataReporter(ws, key, validator, oracle)
		require.Equal(t, key, dp.key)
		require.Equal(t, validator, dp.validatorAccount)
		require.Equal(t, ws, dp.autonityWSUrl)
		require.Equal(t, oracle, dp.oracleService)
	})

	t.Run("gc round data", func(t *testing.T) {
		dp := NewDataReporter(ws, key, validator, oracle)
		for r := 0; r < 100; r++ {
			dp.currentRound = uint64(r)
			var roundData = &types.RoundData{}
			dp.roundData[uint64(r)] = roundData
		}
		require.Equal(t, 100, len(dp.roundData))
		dp.gcRoundData()
		require.Equal(t, MaxBufferedRounds, len(dp.roundData))
	})
}
