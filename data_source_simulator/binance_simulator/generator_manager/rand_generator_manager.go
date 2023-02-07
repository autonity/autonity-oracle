package generator_manager

import (
	"autonity-oracle/data_source_simulator"
	"autonity-oracle/data_source_simulator/binance_simulator/config"
	"autonity-oracle/data_source_simulator/binance_simulator/types"
	"autonity-oracle/data_source_simulator/generators"
	"encoding/csv"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"os"
	"sync"
	"time"
)

var (
	DataGenInterval = 5 * time.Second
)

type RandGeneratorManager struct {
	logger       hclog.Logger
	conf         map[string]*config.RandGeneratorConfig
	mutex        sync.RWMutex
	prices       map[string]decimal.Decimal
	generators   map[string]data_source_simulator.DataGenerator
	dataPointLog string
	symbols      []string
	doneCh       chan struct{}
	jobTicker    *time.Ticker
	file         *os.File
	writer       *csv.Writer
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
		gm.symbols = append(gm.symbols, k)
	}

	err := gm.createDataPointLog()
	if err != nil {
		panic(err)
	}

	gm.logger = hclog.New(&hclog.LoggerOptions{
		Name:   "BinanceSimulator-Random",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})
	return gm
}

func (gm *RandGeneratorManager) createDataPointLog() error {
	// create data point log and write header
	gm.dataPointLog = fmt.Sprintf(".data-point-%d.csv", os.Getpid())
	f, err := os.Create(gm.dataPointLog)
	if err != nil {
		panic(err)
	}

	gm.file = f
	gm.writer = csv.NewWriter(gm.file)
	defer gm.writer.Flush()
	if err := gm.writer.Write(gm.symbols); err != nil {
		panic(err)
	}
	return nil
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
	// record data point record in log file.
	var rec []string
	for _, s := range gm.symbols {
		rec = append(rec, gm.prices[s].String())
	}

	if err := gm.writer.Write(rec); err != nil {
		panic(err)
	}
	defer gm.writer.Flush()
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
	gm.writer.Flush()
	gm.file.Close()
	gm.logger.Info("Data point logs saved at: ", gm.dataPointLog)
}
