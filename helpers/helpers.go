package helpers

import "strings"

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
