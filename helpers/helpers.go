package helpers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"io/fs"
	"io/ioutil" //nolint
	"math/big"
	"os"
	"sort"
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

// Median return the median value in the provided data set
func Median(prices []decimal.Decimal) (decimal.Decimal, error) {
	l := len(prices)
	if l == 0 {
		return decimal.Decimal{}, fmt.Errorf("empty data set for median aggregation")
	}

	if l == 1 {
		return prices[0], nil
	}

	sort.SliceStable(prices, func(i, j int) bool {
		return prices[i].Cmp(prices[j]) == -1
	})

	if len(prices)%2 == 0 {
		return prices[l/2-1].Add(prices[l/2]).Div(decimal.RequireFromString("2.0")), nil
	}

	return prices[l/2], nil
}

// VWAP computes the volume weighted average price for the input prices with their corresponding volumes
func VWAP(prices []decimal.Decimal, volumes []*big.Int) (decimal.Decimal, error) {
	if len(prices) == 0 || len(volumes) == 0 || len(prices) != len(volumes) {
		return decimal.Zero, errors.New("prices and volumes must be of the same non-zero length")
	}

	var totalWeightedPrice decimal.Decimal
	totalVolume := big.NewInt(0)

	for i := range prices {
		// Convert volume to decimal.Decimal
		volumeDecimal := decimal.NewFromBigInt(volumes[i], 0)

		// Calculate weighted price for current price and volume
		weightedPrice := prices[i].Mul(volumeDecimal) // Use decimal.Decimal for precision
		totalWeightedPrice = totalWeightedPrice.Add(weightedPrice)

		// Accumulate total volume
		totalVolume.Add(totalVolume, volumes[i])
	}

	// Avoid division by zero
	if totalVolume.Cmp(big.NewInt(0)) == 0 {
		return decimal.Zero, errors.New("total volume cannot be zero")
	}

	// Calculate VWAP
	vwap := totalWeightedPrice.Div(decimal.NewFromBigInt(totalVolume, 0))

	return vwap, nil
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
