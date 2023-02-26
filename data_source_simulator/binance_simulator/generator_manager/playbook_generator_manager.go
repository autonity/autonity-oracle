package generator_manager

import (
	"autonity-oracle/data_source_simulator/binance_simulator/types"
	"autonity-oracle/helpers"
	"encoding/csv"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	DataPlayInterval = 5 * time.Second
)

type PlaybookGeneratorManager struct {
	logger      hclog.Logger
	playbook    string
	mutex       sync.RWMutex
	prices      map[string]decimal.Decimal
	columnIndex map[int]string
	doneCh      chan struct{}
	jobTicker   *time.Ticker
}

func NewPlaybookGeneratorManager(playbook string) *PlaybookGeneratorManager {
	pm := &PlaybookGeneratorManager{
		playbook:    playbook,
		doneCh:      make(chan struct{}),
		jobTicker:   time.NewTicker(DataPlayInterval),
		prices:      make(map[string]decimal.Decimal),
		columnIndex: make(map[int]string),
	}
	pm.logger = hclog.New(&hclog.LoggerOptions{
		Name:   "BinanceSimulator-Playbook",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	symbols, err := helpers.ParsePlaybookHeader(playbook)
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range symbols {
		pm.columnIndex[i] = s
	}

	return pm
}

func (pm *PlaybookGeneratorManager) GetSymbolPrice(symbols []string) (types.Prices, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	var result types.Prices
	if len(symbols) != 0 {
		for _, s := range symbols {
			if p, ok := pm.prices[s]; !ok {
				return result, fmt.Errorf("InvalidSymbols")
			} else {
				result = append(result, types.Price{
					Symbol: s,
					Price:  p.String(),
				})
			}
		}
	} else {
		for s, p := range pm.prices {
			result = append(result, types.Price{
				Symbol: s,
				Price:  p.String(),
			})
		}
	}

	return result, nil
}

func (pm *PlaybookGeneratorManager) AdjustParams(params types.GeneratorParams, method string) error {
	// not needed for playbook generator.
	return nil
}

func (pm *PlaybookGeneratorManager) UpdatePrices(csvReader *csv.Reader) error {
	rec, err := csvReader.Read()
	if err == io.EOF {
		return err
	}

	if err != nil {
		return err
	}
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	for i, data := range rec {
		p, err := decimal.NewFromString(data)
		if err != nil {
			pm.logger.Error("symbol have invalid value at playbook: ", pm.columnIndex[i], data)
			continue
		}
		pm.logger.Debug("replay data point for symbol: ", pm.columnIndex[i], p.String())
		pm.prices[pm.columnIndex[i]] = p
	}
	return nil
}

func (pm *PlaybookGeneratorManager) Start() {
	f, err := os.Open(pm.playbook)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	headers, err := csvReader.Read()
	if err != nil {
		panic(err)
	}

	pm.logger.Info("running playbook with symbols: ", headers)

	for {
		select {
		case <-pm.doneCh:
			pm.jobTicker.Stop()
			pm.logger.Info("the jobTicker jobs of binance_simulator simulator is stopped")
			return
		case <-pm.jobTicker.C:
			err := pm.UpdatePrices(csvReader)
			if err != nil {
				pm.logger.Error(err.Error())
				pm.jobTicker.Stop()
				return
			}
		}
	}
}

func (pm *PlaybookGeneratorManager) Stop() {
	pm.doneCh <- struct{}{}
}
