package server

import (
	"autonity-oracle/config"
	"autonity-oracle/types"
	"context"

	"github.com/ethereum/go-ethereum/event"
)

func (os *Server) WatchSampleEvent(sink chan<- *types.SampleEvent) event.Subscription {
	return os.sampleEventFeed.Subscribe(sink)
}

func (os *Server) samplingFirstRound(ts int64) error {
	nextRoundHeight := os.votePeriod
	curHeight, err := os.client.BlockNumber(context.Background())
	if err != nil {
		os.logger.Error("handle pre-sampling", "error", err.Error())
		return err
	}

	if curHeight > nextRoundHeight {
		return nil
	}

	if nextRoundHeight-curHeight > uint64(config.PreSamplingRange) { //nolint
		return nil
	}

	// do the data pre-sampling.
	os.logger.Debug("data pre-sampling", "round", os.curRound, "on height", curHeight, "TS", ts)
	os.samplePrice(os.samplingSymbols, ts)
	return nil
}

func (os *Server) handlePreSampling(preSampleTS int64) error {

	// start to sample data point for the 1st round as the round period could be longer than 30s, we don't want to
	// wait for another round to get the data be available on-chain.
	if os.curRound == FirstRound {
		return os.samplingFirstRound(preSampleTS)
	}

	// if it is not a good timing to start sampling then return.
	nextRoundHeight := os.curRoundHeight + os.votePeriod
	curHeight, err := os.client.BlockNumber(context.Background())
	if err != nil {
		os.logger.Error("handle pre-sampling", "error", err.Error())
		return err
	}
	if nextRoundHeight-curHeight > uint64(config.PreSamplingRange) { //nolint
		return nil
	}

	// do the data pre-sampling.
	os.logger.Debug("data pre-sampling", "round", os.curRound, "on height", curHeight, "TS", preSampleTS)
	os.samplePrice(os.samplingSymbols, preSampleTS)
	return nil
}

func (os *Server) samplePrice(symbols []string, ts int64) {
	if os.lastSampledTS == ts {
		return
	}
	cpSymbols := make([]string, len(symbols))
	copy(cpSymbols, symbols)
	e := &types.SampleEvent{
		Symbols: cpSymbols,
		TS:      ts,
	}
	nListener := os.sampleEventFeed.Send(e)
	os.logger.Debug("sample event is sent to", "num of plugins", nListener)
}
