package helpers

import (
	"autonity-oracle/types"
	"encoding/csv"
	"fmt"
	"github.com/namsral/flag"
	"github.com/shopspring/decimal"
	"io"
	"io/fs"
	"io/ioutil" //nolint
	"os"
	"sort"
)

var (
	pEURUSD  = decimal.RequireFromString("1.086")
	pJPYUSD  = decimal.RequireFromString("0.0073")
	pGBPUSD  = decimal.RequireFromString("1.25")
	pAUDUSD  = decimal.RequireFromString("0.67")
	pCADUSD  = decimal.RequireFromString("0.74")
	pSEKUSD  = decimal.RequireFromString("0.096")
	pATNUSD  = decimal.RequireFromString("1.0")
	pUSDCUSD = decimal.RequireFromString("1.0")
	pNTNUSD  = decimal.RequireFromString("10.0")
)

func PrintUsage() {
	fmt.Print("Usage of Autonity Oracle Server:\n")
	fmt.Print("Sub commands: \n  version: print the version of the oracle server.\n")
	fmt.Print("Flags:\n")
	flag.PrintDefaults()
}

func ResolveSimulatedPrice(s string) decimal.Decimal {
	defaultPrice := types.SimulatedPrice
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

func ListPlugins(path string) ([]fs.FileInfo, error) {
	var plugins []fs.FileInfo
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		f := file
		if f.IsDir() {
			continue
		}
		// only executable binaries are returned.
		if !IsExecOwnerGroup(f.Mode()) {
			continue
		}

		plugins = append(plugins, f)
	}
	return plugins, nil
}

// IsExecOwnerGroup return if the file is executable for the owner and the group
func IsExecOwnerGroup(mode os.FileMode) bool {
	return mode&0110 == 0110
}
