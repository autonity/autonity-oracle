package helpers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

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
