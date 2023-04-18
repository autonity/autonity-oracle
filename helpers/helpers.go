package helpers

import (
	"autonity-oracle/types"
	"encoding/csv"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"os"
	"sort"
	"strings"
)

var (
	pNTNUSD = decimal.RequireFromString("7.0")
	pNTNAUD = decimal.RequireFromString("9.856")
	pNTNCAD = decimal.RequireFromString("9.331")
	pNTNEUR = decimal.RequireFromString("6.8369")
	pNTNGBP = decimal.RequireFromString("5.691")
	pNTNJPY = decimal.RequireFromString("897.435")
	pNTNSEK = decimal.RequireFromString("72.163")
)

func ResolveSimulatedPrice(s string) decimal.Decimal {
	defaultPrice := types.SimulatedPrice
	switch s {
	case "NTNUSD": //nolint
		defaultPrice = pNTNUSD
	case "NTNAUD":
		defaultPrice = pNTNAUD
	case "NTNCAD":
		defaultPrice = pNTNCAD
	case "NTNEUR":
		defaultPrice = pNTNEUR
	case "NTNGBP":
		defaultPrice = pNTNGBP
	case "NTNJPY":
		defaultPrice = pNTNJPY
	case "NTNSEK": //nolint
		defaultPrice = pNTNSEK
	}
	return defaultPrice
}

func ParseSymbols(symbols string) []string {
	var symbolArray []string
	syms := strings.Split(symbols, ",")
	for _, s := range syms {
		symbol := strings.TrimSpace(s)
		if len(symbol) == 0 {
			continue
		}
		symbolArray = append(symbolArray, symbol)
	}
	return symbolArray
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
		return decimal.Decimal{}, fmt.Errorf("empty data set")
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
