package generator_manager

import (
	"autonity-oracle/data_source_simulator"
	"autonity-oracle/data_source_simulator/binance_simulator/config"
	"autonity-oracle/data_source_simulator/binance_simulator/types"
	"autonity-oracle/data_source_simulator/generators"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"os"
	"sync"
	"time"
)

var (
	DataGenInterval = 1 * time.Second
)

type RandGeneratorManager struct {
	logger     hclog.Logger
	conf       map[string]*config.RandGeneratorConfig
	mutex      sync.RWMutex
	prices     map[string]decimal.Decimal
	generators map[string]data_source_simulator.DataGenerator
	doneCh     chan struct{}
	jobTicker  *time.Ticker
}

func NewRandGeneratorManager(conf map[string]*config.RandGeneratorConfig) *RandGeneratorManager {
	gm := &RandGeneratorManager{
		conf:       conf,
		doneCh:     make(chan struct{}),
		jobTicker:  time.NewTicker(DataGenInterval),
		prices:     make(map[string]decimal.Decimal),
		generators: make(map[string]data_source_simulator.DataGenerator),
	}
	for k, v := range conf {
		gm.generators[k] = generators.NewRandDataGenerator(v.ReferenceDataPoint, v.DistributionRate)
	}

	gm.logger = hclog.New(&hclog.LoggerOptions{
		Name:   "BinanceSimulator-Random",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})
	return gm
}

func (gm *RandGeneratorManager) GetSymbolPrice(symbols []string) (types.Prices, error) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()
	var result types.Prices
	for _, s := range symbols {
		if p, ok := gm.prices[s]; !ok {
			return result, fmt.Errorf("InvalidSymbols")
		} else {
			result = append(result, types.Price{
				Symbol: s,
				Price:  p.String(),
			})
		}
	}
	return result, nil
}

func (gm *RandGeneratorManager) AdjustParams(params types.GeneratorParams, method string) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	for _, v := range params {
		if _, ok := gm.generators[v.Symbol]; !ok {
			return fmt.Errorf("InavlidSymbol")
		}

		switch method {
		case "move_to":
			gm.generators[v.Symbol].MoveTo(decimal.NewFromFloat(v.Value))
		case "move_by":
			gm.generators[v.Symbol].MoveBy(decimal.NewFromFloat(v.Value))
		case "set_distribution_rate":
			gm.generators[v.Symbol].SetDistributionRate(decimal.NewFromFloat(v.Value))
		default:
			return fmt.Errorf("InvalidMeothd")
		}
	}
	return nil
}

func (gm *RandGeneratorManager) UpdatePrices() {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	for k, gen := range gm.generators {
		g := gen
		gm.prices[k] = g.NextDataPoint()
		gm.logger.Debug("simulator generates price: ", k, gm.prices[k].String())
	}
}

func (gm *RandGeneratorManager) Start() {
	for {
		select {
		case <-gm.doneCh:
			gm.jobTicker.Stop()
			gm.logger.Info("the jobTicker jobs of binance_simulator simulator is stopped")
			return
		case <-gm.jobTicker.C:
			gm.UpdatePrices()
		}
	}
}

func (gm *RandGeneratorManager) Stop() {
	gm.doneCh <- struct{}{}
}
