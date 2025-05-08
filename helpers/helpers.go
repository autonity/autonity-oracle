package helpers

import (
	"autonity-oracle/types"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"io/fs"
	"io/ioutil" //nolint
	"math/big"
	"os"
)

var (
	pEURUSD        = decimal.RequireFromString("1.086")
	pJPYUSD        = decimal.RequireFromString("0.0073")
	pGBPUSD        = decimal.RequireFromString("1.25")
	pAUDUSD        = decimal.RequireFromString("0.67")
	pCADUSD        = decimal.RequireFromString("0.74")
	pSEKUSD        = decimal.RequireFromString("0.096")
	pATNUSD        = decimal.RequireFromString("1.0")
	pUSDCUSD       = decimal.RequireFromString("1.0")
	pNTNUSD        = decimal.RequireFromString("10.0")
	pBTCETH        = decimal.RequireFromString("29.41")
	simulatedPrice = decimal.RequireFromString("11.11")
	SymbolBTCETH   = "BTC-ETH" // for simulation and tests only.
	DefaultSymbols = []string{"AUD-USD", "CAD-USD", "EUR-USD", "GBP-USD", "JPY-USD", "SEK-USD", "ATN-USD", "NTN-USD", "NTN-ATN"}
)

// ResolveSimulatedPrice only for e2e testing.
func ResolveSimulatedPrice(s string) decimal.Decimal {
	defaultPrice := simulatedPrice
	switch s {
	case "EUR-USD":
		defaultPrice = pEURUSD
	case "JPY-USD":
		defaultPrice = pJPYUSD
	case "GBP-USD":
		defaultPrice = pGBPUSD
	case "AUD-USD":
		defaultPrice = pAUDUSD
	case "CAD-USD":
		defaultPrice = pCADUSD
	case "SEK-USD":
		defaultPrice = pSEKUSD
	case "ATN-USDC": //nolint
		defaultPrice = pATNUSD
	case "ATN-USD": //nolint
		defaultPrice = pATNUSD
	case "NTN-USD": //nolint
		defaultPrice = pNTNUSD
	case "NTN-USDC": //nolint
		defaultPrice = pNTNUSD
	case "USDC-USD":
		defaultPrice = pUSDCUSD
	case SymbolBTCETH:
		defaultPrice = pBTCETH
	}
	return defaultPrice
}

// ResolveSimulatedVolume only for e2e testing.
func ResolveSimulatedVolume(s string) *big.Int {
	defaultVolume := types.NoVolumeData
	initialVolume := new(big.Int).SetUint64(100)
	switch s {
	case "ATN-USDC": //nolint
		defaultVolume = initialVolume
	case "NTN-USDC": //nolint
		defaultVolume = initialVolume
	case "NTN-ATN": //nolint
		defaultVolume = initialVolume
	}
	return defaultVolume
}

func ParsePlaybookHeader(playbook string) ([]string, error) {
	f, err := os.Open(playbook)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	rec, err := csvReader.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("empty playbook file")
	}
	if err != nil {
		return nil, err
	}

	return rec, nil
}

// XWAP computes the X(volume/confidence) weighted average price for the input prices with their corresponding weights
func XWAP(prices []decimal.Decimal, weights []*big.Int) (decimal.Decimal, *big.Int, error) {
	if len(prices) == 0 || len(weights) == 0 || len(prices) != len(weights) {
		return decimal.Zero, nil, errors.New("prices and weights must be of the same non-zero length")
	}

	var totalValue decimal.Decimal
	totalWeights := big.NewInt(0)

	for i := range prices {
		volume := decimal.NewFromBigInt(weights[i], 0)
		value := prices[i].Mul(volume)
		totalValue = totalValue.Add(value)

		// Accumulate total volume
		totalWeights.Add(totalWeights, weights[i])
	}

	// Avoid division by zero
	if totalWeights.Cmp(big.NewInt(0)) == 0 {
		return decimal.Zero, nil, errors.New("total volume cannot be zero")
	}

	// Calculate XWAP
	vwap := totalValue.Div(decimal.NewFromBigInt(totalWeights, 0))
	return vwap, totalWeights, nil
}

// ListPlugins returns a mapping of file names to fs.FileInfo for executable files in the specified path.
func ListPlugins(path string) (map[string]fs.FileInfo, error) {
	plugins := make(map[string]fs.FileInfo)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// Only executable binaries are returned.
		if !IsExecOwnerGroup(file.Mode()) {
			continue
		}

		plugins[file.Name()] = file
	}
	return plugins, nil
}

// IsExecOwnerGroup return if the file is executable for the owner and the group
func IsExecOwnerGroup(mode os.FileMode) bool {
	return mode&0110 == 0110
}
