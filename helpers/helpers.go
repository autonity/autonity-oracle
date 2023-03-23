package helpers

import (
	"autonity-oracle/types"
	"encoding/csv"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"os"
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
	case "NTNUSD":
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
	case "NTNSEK":
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
