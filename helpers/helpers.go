package helpers

import (
	"autonity-oracle/types"
	"encoding/csv"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"io/fs"
	"io/ioutil" //nolint
	"log"
	"os"
	"sort"
	"strings"
)

var (
	pEURUSD = decimal.RequireFromString("1.086")
	pJPYUSD = decimal.RequireFromString("0.0073")
	pGBPUSD = decimal.RequireFromString("1.25")
	pAUDUSD = decimal.RequireFromString("0.67")
	pCADUSD = decimal.RequireFromString("0.74")
	pSEKUSD = decimal.RequireFromString("0.096")
	pATNUSD = decimal.RequireFromString("1.0")
	pNTNUSD = decimal.RequireFromString("10.0")
)

func PrintUsage() {
	usage := "Usage of ./autoracle:\n" +
		"Sub commands: \n" +
		"\tversion: prints the version of the oracle server.\n" +
		"Flags: \n" +
		"\t-key.password=\"123\": set the password to the oracle server key file.\n\n" +
		"\t-plugin.dir=\"./build/bin/plugins\": set the directory of the adapter plugins.\n\n" +
		"\t-gas.tip.cap=1: set the gas priority fee cap to issue the oracle data report transactions.\n\n" +
		"\t-plugin.conf=\"./build/bin/plugins/plugins-conf.yml\": set the plugins' configuration file.\n\n" +
		"\t-log.level=2: Set the logging level, available levels are:  0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error.\n\n" +
		"\t-symbols=\"AUD-USD,CAD-USD,EUR-USD,GBP-USD,JPY-USD,SEK-USD,ATN-USD,NTN-USD,NTN-ATN\": set the symbols string separated by comma.\n\n" +
		"\t-autonity.ws.url=\"ws://127.0.0.1:8546\": set the WS-RPC server listening interface and port of the connected Autonity client node.\n\n" +
		"\t-key.file=\"./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe\": set the oracle server key file.\n"
	log.Print(usage)
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
	case "ATN-USD": //nolint
		defaultPrice = pATNUSD
	case "NTN-USD": //nolint
		defaultPrice = pNTNUSD
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
