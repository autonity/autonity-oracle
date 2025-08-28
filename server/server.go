package server

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	"autonity-oracle/monitor"
	common2 "autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"context"
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	o "os"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
)

var BridgedSymbols = map[string]string{
	ATNUSD: ATNUSDC,
	NTNUSD: NTNUSDC,
}

var (
	saltRange       = new(big.Int).SetUint64(math.MaxInt64)
	alertBalance    = new(big.Int).SetUint64(2000000000000) // 2000 Gwei, 0.000002 Ether
	invalidPrice    = big.NewInt(0)
	invalidSalt     = big.NewInt(0)
	tenSecsInterval = 10 * time.Second // ticker to check L2 connectivity and gc round data.
	oneSecsInterval = 1 * time.Second  // sampling interval during data pre-sampling period.
)

const (
	FirstRound          = uint64(1)
	ATNUSD              = "ATN-USD"
	NTNUSD              = "NTN-USD"
	USDCUSD             = "USDC-USD"
	ATNUSDC             = "ATN-USDC"
	NTNUSDC             = "NTN-USDC"
	MaxConfidence       = 100
	BaseConfidence      = 40
	OracleDecimals      = uint8(18)
	MaxBufferedRounds   = 10
	SourceScalingFactor = uint64(10)
	penalizeEventName   = "Penalized"
)

// Server coordinates the plugin discovery, the data sampling, and do the health checking with L1 connectivity.
type Server struct {
	logger hclog.Logger
	conf   *config.Config

	doneCh        chan struct{}
	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	psTicker      *time.Ticker // the pre-sampling ticker in 1s.

	//runningPlugins  map[string]*pWrapper.PluginWrapper // the plugin clients that connect with different adapters.

	samplingSymbols []string // the symbols for data fetching in oracle service, can be different from the required protocol symbols.

	//keyRequiredPlugins map[string]struct{} // saving those plugins which require a key granted by data provider

	// the reporting staffs
	dialer         types.Dialer
	oracleContract contract.ContractAPI
	client         types.Blockchain
	abi            abi.ABI

	curRound       uint64 //round ID.
	votePeriod     uint64 //vote period.
	curSampleTS    int64  //the data sample TS of the current round.
	curRoundHeight uint64 //The block height on which the last round rotation happens.

	protocolSymbols []string //symbols required for the voting on the oracle contract protocol.
	pricePrecision  decimal.Decimal

	voteRecords VoteRecords

	chInvalidVote  chan *contract.OracleInvalidVote
	subInvalidVote event.Subscription

	chVotedEvent  chan *contract.OracleSuccessfulVote
	subVotedEvent event.Subscription

	chRewardEvent  chan *contract.OracleTotalOracleRewards
	subRewardEvent event.Subscription

	chNoRevealEvent  chan *contract.OracleNoRevealPenalty
	subNoRevealEvent event.Subscription

	chPenalizedEvent  chan *contract.OraclePenalized
	subPenalizedEvent event.Subscription

	chRoundEvent  chan *contract.OracleNewRound
	subRoundEvent event.Subscription

	chSymbolsEvent  chan *contract.OracleNewSymbols
	subSymbolsEvent event.Subscription
	lastSampledTS   int64

	sampleEventFeed        event.Feed
	lostSync               bool // set to true if the connectivity with L1 Autonity network is dropped during runtime.
	commitmentHashComputer *CommitmentHashComputer

	memories Memories

	pluginManager *PluginManager
	//configWatcher  *fsnotify.Watcher // config file watcher which watches the config changes.
	//pluginsWatcher *fsnotify.Watcher // plugins watcher which watches the changes of plugins and the plugins' configs.
	//chainID        int64             // ChainID saves the L1 chain ID, it is used for plugin compatibility check.
}

func NewServer(conf *config.Config, dialer types.Dialer, client types.Blockchain,
	oc contract.ContractAPI) *Server {
	os := &Server{
		conf:           conf,
		dialer:         dialer,
		client:         client,
		oracleContract: oc,
		voteRecords:    make(map[uint64]*types.VoteRecord),
		//runningPlugins:     make(map[string]*pWrapper.PluginWrapper),
		//keyRequiredPlugins: make(map[string]struct{}),
		doneCh:         make(chan struct{}),
		regularTicker:  time.NewTicker(tenSecsInterval),
		psTicker:       time.NewTicker(oneSecsInterval),
		pricePrecision: decimal.NewFromBigInt(common.Big1, int32(OracleDecimals)),
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String() + conf.Key.Address.String(),
		Output: o.Stdout,
		Level:  conf.LoggingLevel,
	})

	abi, err := abi.JSON(strings.NewReader(contract.OracleMetaData.ABI))
	if err != nil {
		os.logger.Error("failed to load ABI", "error", err)
		o.Exit(1)
	}
	os.abi = abi

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		os.logger.Error("failed to get chain id", "error", err)
		o.Exit(1)
	}

	//os.chainID = chainID.Int64()

	commitmentHashComputer, err := NewCommitmentHashComputer()
	if err != nil {
		os.logger.Error("cannot create commitment hash computer", "err", err)
		o.Exit(1)
	}
	os.commitmentHashComputer = commitmentHashComputer

	// load memories from persistence.
	os.memories = Memories{dataDir: conf.ProfileDir}
	os.memories.init(os.logger)
	if os.memories.voteRecords != nil {
		os.voteRecords = *os.memories.voteRecords
		os.logger.Info("loaded vote records from persistence", "records", len(os.voteRecords))
	}

	/*
		// discover plugins from plugin dir at startup.
		binaries, err := helpers.ListPlugins(conf.PluginDir)
		if len(binaries) == 0 || err != nil {
			// to stop the service on the start once there is no plugin in the db.
			os.logger.Error("no plugin discovered", "plugin-dir", os.conf.PluginDir)
			o.Exit(1)
		}

		// take the custom plugin configs and start plugins.
		for _, file := range binaries {
			f := file
			pConf := conf.PluginConfigs[f.Name()]
			if pConf.Disabled {
				continue
			}
			os.tryToLaunchPlugin(f, pConf)
		}*/

	os.logger.Info("running oracle contract listener at", "WS", conf.AutonityWSUrl, "ID", conf.Key.Address.String())
	err = os.sync()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		os.logger.Error("Cannot synchronize oracle contract state from Autonity L1 Network", "error", err.Error(), "WS", conf.AutonityWSUrl)
		o.Exit(1)
	}
	os.lostSync = false

	os.pluginManager = NewPluginManager(os.conf.ConfigFile, os.conf.PluginDir, os.conf.LoggingLevel, os, chainID.Int64(),
		os.conf.Key.Address, os.conf.PluginConfigs)

	/*
		// subscribe FS notifications of the watched plugins.
		pluginsWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			os.logger.Error("cannot create fsnotify watcher", "error", err)
			o.Exit(1)
		}

		err = pluginsWatcher.Add(conf.PluginDir)
		if err != nil {
			os.logger.Error("cannot watch plugin dir", "error", err)
			o.Exit(1)
		}
		os.pluginsWatcher = pluginsWatcher
	*/

	/*
		// subscribe FS notification of the watched config file.
		configWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			os.logger.Error("cannot create fsnotify watcher", "error", err)
			o.Exit(1)
		}

		dir := filepath.Dir(conf.ConfigFile)
		err = configWatcher.Add(dir) // Watch parent directory
		if err != nil {
			os.logger.Error("cannot watch oracle config directory", "error", err)
			o.Exit(1)
		}

		os.configWatcher = configWatcher
	*/
	return os
}

func (os *Server) WatchSampleEvent(sink chan<- *types.SampleEvent) event.Subscription {
	return os.sampleEventFeed.Subscribe(sink)
}

func (os *Server) Start() {
	for {
		select {
		case <-os.doneCh:
			os.regularTicker.Stop()
			os.psTicker.Stop()
			/*
				if os.pluginsWatcher != nil {
					os.pluginsWatcher.Close() //nolint
				}

				if os.configWatcher != nil {
					os.configWatcher.Close() //nolint
				}*/
			os.logger.Info("oracle service is stopped")
			return
		/*
			case err := <-os.configWatcher.Errors:
				if err != nil {
					os.logger.Error("oracle config file watcher err", "err", err.Error())
				}
			case err := <-os.pluginsWatcher.Errors:
				if err != nil {
					os.logger.Error("plugin watcher errors", "err", err.Error())
				}
		*/
		case err := <-os.subSymbolsEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new symbols event", err)
				os.handleConnectivityError()
				os.subSymbolsEvent.Unsubscribe()
			}
		case err := <-os.subRoundEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new round event", err)
				os.handleConnectivityError()
				os.subRoundEvent.Unsubscribe()
			}
		case err := <-os.subRewardEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of reward event", err)
				os.handleConnectivityError()
				os.subRewardEvent.Unsubscribe()
			}
		case err := <-os.subInvalidVote.Err():
			if err != nil {
				os.logger.Info("subscription error of invalid vote", err)
				os.handleConnectivityError()
				os.subInvalidVote.Unsubscribe()
			}
		case err := <-os.subVotedEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of voted event", err)
				os.handleConnectivityError()
				os.subVotedEvent.Unsubscribe()
			}
		case err := <-os.subNoRevealEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of no reveal event", err)
				os.handleConnectivityError()
				os.subNoRevealEvent.Unsubscribe()
			}
		case err := <-os.subPenalizedEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of penalty event", err)
				os.handleConnectivityError()
				os.subPenalizedEvent.Unsubscribe()
			}
		case <-os.psTicker.C:
			// shorten the health checker, if we have L1 connectivity issue, try to repair it before pre sampling starts.
			os.checkHealth()
			preSampleTS := time.Now().Unix()
			err := os.handlePreSampling(preSampleTS)
			if err != nil {
				os.logger.Error("handle pre-sampling", "error", err.Error())
			}
			os.lastSampledTS = preSampleTS

		case invalidVote := <-os.chInvalidVote:
			os.logger.Info("received invalid vote", "cause", invalidVote.Cause, "expected",
				invalidVote.ExpValue.String(), "actual", invalidVote.ActualValue.String(), "txn", invalidVote.Raw.TxHash)
			os.setVoteMined(invalidVote.Raw.TxHash, invalidVote.Cause)
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.InvalidVoteMetric, nil).Inc(1)
			}

		case votedEvent := <-os.chVotedEvent:
			os.logger.Info("received voted event", "height", votedEvent.Raw.BlockNumber, "txn", votedEvent.Raw.TxHash)
			os.setVoteMined(votedEvent.Raw.TxHash, "")
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.SuccessfulVoteMetric, nil).Inc(1)
			}

		case rewardEvent := <-os.chRewardEvent:
			os.logger.Info("received reward distribution event", "height", rewardEvent.Raw.BlockNumber,
				"total distributed ATN", rewardEvent.AtnReward.Uint64(), "total distributed NTN", rewardEvent.NtnReward.Uint64())

		case noRevealEvent := <-os.chNoRevealEvent:
			os.logger.Info("received no reveal event", "height", noRevealEvent.Raw.BlockNumber, "round", noRevealEvent.Round,
				"missed", noRevealEvent.MissedReveal.Uint64())
			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.NoRevealVoteMetric, nil).Update(noRevealEvent.MissedReveal.Int64())
			}

		case penalizeEvent := <-os.chPenalizedEvent:
			if err := os.handlePenaltyEvent(penalizeEvent); err != nil {
				os.logger.Error("handle penalty event", "error", err.Error())
			}

		/*
			case fsEvent, ok := <-os.configWatcher.Events:
				if !ok {
					os.logger.Error("config watcher channel has been closed")
					return
				}
				os.handleConfigEvent(fsEvent)

			case fsEvent, ok := <-os.pluginsWatcher.Events:
				if !ok {
					os.logger.Error("plugin watcher channel has been closed")
					return
				}

				os.logger.Info("watched plugins fs event", "file", fsEvent.Name, "event", fsEvent.Op.String())
				// updates on the watched plugin directory will trigger plugin management.
				os.PluginRuntimeManagement()
		*/
		case roundEvent := <-os.chRoundEvent:
			os.logger.Info("handle new round", "round", roundEvent.Round.Uint64(), "required sampling TS",
				roundEvent.Timestamp.Uint64(), "height", roundEvent.Raw.BlockNumber, "round period", roundEvent.VotePeriod.Uint64())

			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.RoundMetric, nil).Update(roundEvent.Round.Int64())
			}

			// IMPORTANT! sync the round and vote period carries by the round event, and store the reference points for
			// next presampling and the target for sample selection.
			os.curRound = roundEvent.Round.Uint64()
			os.votePeriod = roundEvent.VotePeriod.Uint64()
			os.curRoundHeight = roundEvent.Raw.BlockNumber
			os.curSampleTS = roundEvent.Timestamp.Int64()

			// vote for latest protocol symbols.
			err := os.vote()
			if err != nil {
				os.logger.Error("round voting failed", "error", err.Error())
			}
			// after vote, reset sampling symbols with the latest protocol symbols.
			os.resetSamplingSymbols(os.protocolSymbols)
			os.printLatestRoundData(os.curRound)
			os.gcStaleSamples()
		case newSymbolEvent := <-os.chSymbolsEvent:
			// New symbols are added, add them into the sampling set to prepare data in advance for the coming round's vote.
			os.logger.Info("handle new symbols", "new symbols", newSymbolEvent.Symbols, "activate at round", newSymbolEvent.Round)
			os.handleNewSymbolsEvent(newSymbolEvent.Symbols)
		case <-os.regularTicker.C:
			os.trackVoteState()
			os.gcVoteRecords()
			/*
				if metrics.Enabled {
					metrics.GetOrRegisterGauge(monitor.PluginMetric, nil).Update(int64(len(os.runningPlugins)))
				}*/
			os.logger.Debug("current round ID", "round", os.curRound)
		}
	}
}

func (os *Server) Stop() {
	os.client.Close()
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()
	os.subPenalizedEvent.Unsubscribe()
	os.subNoRevealEvent.Unsubscribe()
	os.subVotedEvent.Unsubscribe()
	os.subInvalidVote.Unsubscribe()
	os.subRewardEvent.Unsubscribe()

	os.doneCh <- struct{}{}
	if os.pluginManager != nil {
		os.pluginManager.Stop()
	}
}

// trackVoteState works in a pull mode to track if the vote was mined by L1 although there is already a push mode
// which subscribe the events from oracle contract. We cannot sure that L1 node is already on service, for example,
// some operation, synchronization, resetting, etc...
func (os *Server) trackVoteState() {

	var update bool
	for r := os.curRound; r > os.curRound-MaxBufferedRounds; r-- {
		vote, ok := os.voteRecords[r]
		if !ok {
			continue
		}

		if vote.Mined {
			continue
		}

		receipt, err := os.client.TransactionReceipt(context.Background(), vote.TxHash)
		if err != nil {
			os.logger.Info("cannot get vote receipt yet", "txn", vote.TxHash, "error", err.Error())
			continue
		}

		vote.Mined = true
		update = true
		os.logger.Info("last vote get mined", "txn", vote.TxHash, "receipt", receipt)
	}

	if update {
		if err := os.memories.flushRecord(os.voteRecords); err != nil {
			os.logger.Warn("failed to flush vote record to persistence", "error", err.Error())
			os.logger.Warn("IMPORTANT: please check your profile data dir, the server need to flush vote record into it.")
		}
	}
}

func (os *Server) setVoteMined(hash common.Hash, err string) {

	var update bool
	// iterate from the most recent round.
	for r := os.curRound; r > os.curRound-MaxBufferedRounds; r-- {
		if vote, ok := os.voteRecords[r]; ok {
			if vote.TxHash == hash {
				if !vote.Mined {
					vote.Mined = true
					vote.Error = err
					update = true
					break
				}
				// not state change, just skip the flushing.
				return
			}
		}
	}

	// flush the change of state
	if update {
		if err := os.memories.flushRecord(os.voteRecords); err != nil {
			os.logger.Warn("failed to flush vote record to persistence", "error", err.Error())
			os.logger.Warn("IMPORTANT: please check your profile data dir, the server need to flush vote record into it.")
		}
		os.logger.Info("vote get mined", "hash", hash, "current round", os.curRound)
		return
	}

	os.logger.Warn("cannot find the round vote with TXN hash", "current round", os.curRound, "hash", hash)
}

/*
func (os *Server) handleConfigEvent(ev fsnotify.Event) {
	// filter unwatched files in the dir.
	if filepath.Base(ev.Name) != filepath.Base(os.conf.ConfigFile) {
		return
	}

	switch {
	// tools like sed issues write event for the updates.
	case ev.Op&fsnotify.Write > 0:
		// apply plugin config changes.
		os.logger.Info("config file content changed", "file", ev.Name)
		os.PluginRuntimeManagement()

	// tools like vim or vscode issues rename, chmod and remove events for the update for an atomic change mode.
	case ev.Op&fsnotify.Rename > 0:
		os.logger.Info("config file changed", "file", ev.Name)
		os.PluginRuntimeManagement()
	}
}*/

func (os *Server) handlePenaltyEvent(penalizeEvent *contract.OraclePenalized) error {
	// As the OutlierDetectionThreshold is set to low level, e.g. 3% against the median, and the OutlierSlashingThreshold
	// is configured at (10%, 15%) which is much higher, a penalization event may occur with zero slashing amount.
	// This indicates that the current client has been identified as an outlier but is not penalized, as its data
	// point falls below the OutlierSlashingThreshold when compared to the median price. To ensure a broader participation
	// of nodes within the oracle network and maintain its operational liveness, we continue to allow these
	// non-slashed outliers to contribute data samples to the network.
	if metrics.Enabled {
		gap := new(big.Int).Abs(new(big.Int).Sub(penalizeEvent.Reported, penalizeEvent.Median))
		gapPercent := new(big.Int).Div(new(big.Int).Mul(gap, big.NewInt(100)), penalizeEvent.Median)
		metrics.GetOrRegisterGauge(monitor.OutlierDistancePercentMetric, nil).Update(gapPercent.Int64())
	}

	if penalizeEvent.SlashingAmount.Cmp(common.Big0) == 0 {
		os.logger.Warn("Client addressed as an outlier, the last vote won't be counted for reward distribution, "+
			"please use high quality data source.", "symbol", penalizeEvent.Symbol, "median value",
			penalizeEvent.Median.String(), "reported value", penalizeEvent.Reported.String())
		os.logger.Warn("IMPORTANT: please double check your data source setup before getting penalized")
		if metrics.Enabled {
			metrics.GetOrRegisterCounter(monitor.OutlierNoSlashTimesMetric, nil).Inc(1)
		}
		return nil
	}

	os.logger.Warn("Client get penalized as an outlier", "node", penalizeEvent.Participant,
		"currency symbol", penalizeEvent.Symbol, "median value", penalizeEvent.Median.String(),
		"reported value", penalizeEvent.Reported.String(), "block", penalizeEvent.Raw.BlockNumber, "slashed amount", penalizeEvent.SlashingAmount.Uint64())
	os.logger.Warn("your next vote will be postponed", "in blocks", os.conf.VoteBuffer)
	os.logger.Warn("IMPORTANT: please repair your data setups for data precision before getting penalized again")

	if metrics.Enabled {
		metrics.GetOrRegisterCounter(monitor.OutlierSlashTimesMetric, nil).Inc(1)
		baseUnitsPerNTN := new(big.Float).SetInt(big.NewInt(1e18))
		amount := new(big.Float).SetUint64(penalizeEvent.SlashingAmount.Uint64())
		ntnFloat, _ := new(big.Float).Quo(amount, baseUnitsPerNTN).Float64()
		metrics.GetOrRegisterGaugeFloat64(monitor.OutlierPenaltyMetric, nil).Update(ntnFloat)
	}

	outlierRecord := &OutlierRecord{
		LastPenalizedAtBlock: penalizeEvent.Raw.BlockNumber,
		Participant:          penalizeEvent.Participant,
		Symbol:               penalizeEvent.Symbol,
		Median:               penalizeEvent.Median.Uint64(),
		Reported:             penalizeEvent.Reported.Uint64(),
		SlashingAmount:       penalizeEvent.SlashingAmount.Uint64(),
		LoggedAt:             time.Now().Format(time.RFC3339),
	}
	os.memories.outlierRecord = outlierRecord
	if err := os.memories.flushRecord(outlierRecord); err != nil {
		os.logger.Warn("failed to flush penality record to persistence", "error", err.Error())
		os.logger.Warn("IMPORTANT: please check your profile data dir, the server need to flush penality record into it.")
		return err
	}
	return nil
}

// sync is executed on client startup or after the L1 connection recovery to sync the on-chain oracle contract
// states, symbols, round id, precision, vote period, etc... to the oracle server. It also subscribes the on-chain
// events of oracle protocol: round event, symbol update event, etc...
func (os *Server) sync() error {
	var err error
	// get initial states from oracle contract.
	os.curRoundHeight, os.curRound, os.protocolSymbols, os.votePeriod, err = os.syncRoundState()
	if err != nil {
		os.logger.Error("synchronize oracle contract state", "error", err.Error())
		return err
	}

	// reset sampling symbols with the latest protocol symbols, it adds bridger symbols by according to the protocol symbols.
	os.resetSamplingSymbols(os.protocolSymbols)

	// subscribe protocol events
	if err = os.subscribeEvents(); err != nil {
		return err
	}
	os.logger.Info("synced", "CurrentRoundHeight", os.curRoundHeight, "CurrentRound", os.curRound,
		"protocol symbols", os.protocolSymbols, "sampling symbols", os.samplingSymbols)
	return nil
}

func (os *Server) subscribeEvents() error {
	// subscribe on-chain round rotation event
	chRoundEvent := make(chan *contract.OracleNewRound)
	subRoundEvent, err := os.oracleContract.WatchNewRound(new(bind.WatchOpts), chRoundEvent)
	if err != nil {
		os.logger.Error("failed to subscribe round event", "error", err.Error())
		return err
	}
	os.chRoundEvent = chRoundEvent
	os.subRoundEvent = subRoundEvent

	// subscribe on-chain symbol update event
	chSymbolsEvent := make(chan *contract.OracleNewSymbols)
	subSymbolsEvent, err := os.oracleContract.WatchNewSymbols(new(bind.WatchOpts), chSymbolsEvent)
	if err != nil {
		os.logger.Error("failed to subscribe new symbol event", "error", err.Error())
		return err
	}
	os.chSymbolsEvent = chSymbolsEvent
	os.subSymbolsEvent = subSymbolsEvent

	// subscribe on-chain no-reveal event
	chNoRevealEvent := make(chan *contract.OracleNoRevealPenalty)
	subNoRevealEvent, err := os.oracleContract.WatchNoRevealPenalty(new(bind.WatchOpts), chNoRevealEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe no reveal event", "error", err.Error())
		return err
	}
	os.chNoRevealEvent = chNoRevealEvent
	os.subNoRevealEvent = subNoRevealEvent

	// subscribe on-chain penalize event
	chPenalizedEvent := make(chan *contract.OraclePenalized)
	subPenalizedEvent, err := os.oracleContract.WatchPenalized(new(bind.WatchOpts), chPenalizedEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe penalized event", "error", err.Error())
		return err
	}
	os.chPenalizedEvent = chPenalizedEvent
	os.subPenalizedEvent = subPenalizedEvent

	// subscribe voted event
	chVotedEvent := make(chan *contract.OracleSuccessfulVote)
	subVotedEvent, err := os.oracleContract.WatchSuccessfulVote(new(bind.WatchOpts), chVotedEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe voted event", "error", err.Error())
		return err
	}
	os.chVotedEvent = chVotedEvent
	os.subVotedEvent = subVotedEvent

	// subscribe invalid vote event
	chInvalidVote := make(chan *contract.OracleInvalidVote)
	subInvalidVote, err := os.oracleContract.WatchInvalidVote(new(bind.WatchOpts), chInvalidVote, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe invalid vote event", "error", err.Error())
		return err
	}
	os.chInvalidVote = chInvalidVote
	os.subInvalidVote = subInvalidVote

	// subscribe reward event
	chRewardEvent := make(chan *contract.OracleTotalOracleRewards)
	subRewardEvent, err := os.oracleContract.WatchTotalOracleRewards(new(bind.WatchOpts), chRewardEvent)
	if err != nil {
		os.logger.Error("failed to subscribe reward event", "error", err.Error())
		return err
	}
	os.chRewardEvent = chRewardEvent
	os.subRewardEvent = subRewardEvent
	return nil
}

// syncRoundState returns round id, symbols and vote period on oracle contract, it is called on the startup of client.
// Since below steps are not atomic get operation from blockchain, thus they are just being used at the initial phase
// for data presampling, the correctness of voting is promised by the synchronization triggered by the round event before
// the voting.
func (os *Server) syncRoundState() (uint64, uint64, []string, uint64, error) {
	// on the startup, we need to sync the round block, round id, symbols and committees from contract.
	currentRoundHeight, err := os.oracleContract.GetLastRoundBlock(nil)
	if err != nil {
		os.logger.Error("get round block", "error", err.Error())
		return 0, 0, nil, 0, err
	}

	currentRound, err := os.oracleContract.GetRound(nil)
	if err != nil {
		os.logger.Error("get round", "error", err.Error())
		return 0, 0, nil, 0, err
	}

	symbols, err := os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("get symbols", "error", err.Error())
		return 0, 0, nil, 0, err
	}

	votePeriod, err := os.oracleContract.GetVotePeriod(nil)
	if err != nil {
		os.logger.Error("get vote period", "error", err.Error())
		return 0, 0, nil, 0, nil
	}

	if len(symbols) == 0 {
		os.logger.Error("there are no symbols in Autonity L1 oracle contract")
		return currentRoundHeight.Uint64(), currentRound.Uint64(), symbols, votePeriod.Uint64(), types.ErrNoSymbolsObserved
	}

	return currentRoundHeight.Uint64(), currentRound.Uint64(), symbols, votePeriod.Uint64(), nil
}

func (os *Server) gcStaleSamples() {
	os.pluginManager.GCSamples()

	/*
		for _, plugin := range os.runningPlugins {
			plugin.GCExpiredSamples()
		}*/
}

func (os *Server) gcVoteRecords() {
	if len(os.voteRecords) >= MaxBufferedRounds {
		offset := os.curRound - MaxBufferedRounds
		for k := range os.voteRecords {
			if k <= offset {
				delete(os.voteRecords, k)
			}
		}
	}
}

func (os *Server) handleConnectivityError() {
	os.lostSync = true
}

func (os *Server) checkHealth() {
	if os.lostSync {
		err := os.sync()
		if err != nil && !errors.Is(err, types.ErrNoSymbolsObserved) {
			os.logger.Info("rebuilding WS connectivity with Autonity L1 node", "error", err)
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.L1ConnectivityMetric, nil).Inc(1)
			}
			return
		}
		os.lostSync = false
		return
	}
}

func (os *Server) isVoter() (bool, error) {
	voters, err := os.oracleContract.GetVoters(nil)
	if err != nil {
		os.logger.Error("get voters", "error", err.Error())
		return false, err
	}

	for _, c := range voters {
		if c == os.conf.Key.Address {
			return true, nil
		}
	}
	return false, nil
}

func (os *Server) printLatestRoundData(newRound uint64) {
	for _, s := range os.protocolSymbols {
		rd, err := os.oracleContract.GetRoundData(nil, new(big.Int).SetUint64(newRound-1), s)
		if err != nil {
			os.logger.Error("get round data", "error", err.Error())
			return
		}

		os.logger.Debug("get round price", "round", newRound-1, "symbol", s, "Price",
			rd.Price.String(), "success", rd.Success)
	}

	for _, s := range os.protocolSymbols {
		rd, err := os.oracleContract.LatestRoundData(nil, s)
		if err != nil {
			os.logger.Error("get latest round price", "error", err.Error())
			return
		}

		price, err := decimal.NewFromString(rd.Price.String())
		if err != nil {
			continue
		}

		os.logger.Debug("latest round price", "round", rd.Round.Uint64(), "symbol", s, "price",
			price.Div(os.pricePrecision).String(), "success", rd.Success)
	}
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

func (os *Server) isBlockchainSynced() bool {
	// if the autonity node is on peer synchronization state, just skip the reporting.
	syncing, err := os.client.SyncProgress(context.Background())
	if err != nil {
		os.logger.Error("vote get SyncProgress", "error", err.Error())
		return false
	}

	if syncing != nil {
		os.logger.Warn("skip round event since the Autonity L1 node is doing block synchronization")
		return false
	}

	return true
}

func (os *Server) syncProtocolSymbols() error {
	// get latest symbols from oracle.
	var err error
	os.protocolSymbols, err = os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("vote get symbols", "error", err.Error())
		return err
	}

	return nil
}

func (os *Server) penaltyTopic(name string, query ...[]interface{}) ([][]common.Hash, error) {
	// Append the event selector to the query parameters and construct the topic set
	query = append([][]interface{}{{os.abi.Events[name].ID}}, query...)
	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// UnpackLog unpacks a retrieved log into the provided output structure.
func (os *Server) unpackLog(out interface{}, event string, log tp.Log) error {
	if len(log.Data) > 0 {
		if err := os.abi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range os.abi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

func (os *Server) checkOutlierSlashing() bool {
	// filer log with the topic of penalized event with self address.
	var participants []interface{}
	participants = append(participants, os.conf.Key.Address)
	topic, err := os.penaltyTopic(penalizeEventName, participants)
	if err != nil {
		os.logger.Error("fail to assemble penality topic", "error", err.Error(), "height", os.curRoundHeight)
		return false
	}

	// filter log over the round block.
	logs, err := os.client.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(os.curRoundHeight),
		ToBlock:   new(big.Int).SetUint64(os.curRoundHeight),
		Addresses: []common.Address{types.OracleContractAddress},
		Topics:    topic,
	})
	if err != nil {
		os.logger.Info("fail to filter logs", "height", os.curRoundHeight, "err", err.Error())
		return false
	}

	// No penalized event at all.
	if len(logs) == 0 {
		return false
	}

	// As the logs are filtered by topic and indexed by the participant address, thus the logged event should be the
	// one we watched unless the L1 client was wrong.
	if len(logs) > 1 {
		// This is not expected unless there is a L1 protocol bug.
		os.logger.Warn("L1 network emits multiple outlier penality events against the client at the end of round")
	}

	log := logs[0]
	ev := new(contract.OraclePenalized)
	if err = os.unpackLog(ev, penalizeEventName, log); err != nil {
		os.logger.Error("failed to unpack outlier penalize event", "error", err, "height", os.curRoundHeight)
		return false
	}

	if ev.SlashingAmount.Cmp(common.Big0) > 0 {
		os.logger.Info("on going slashing", "height", os.curRoundHeight, "round", os.curRound,
			"symbol", ev.Symbol, "median", ev.Median.String(), "reported", ev.Reported.String(), "slashing amount",
			ev.SlashingAmount.String())
		return true
	}
	return false
}

func (os *Server) vote() error {
	if !os.isBlockchainSynced() {
		return types.ErrPeerOnSync
	}

	// sync protocol symbols before vote.
	if err := os.syncProtocolSymbols(); err != nil {
		return err
	}

	// as outlier slashing event can come right after round event in the same block.
	// if node is on outlier slashing, skip round vote to avoid the outlier slashing again.
	if os.checkOutlierSlashing() {
		return types.ErrOnOutlierSlashing
	}

	// if client is not a voter, just skip reporting.
	isVoter, err := os.isVoter()
	if err != nil {
		os.logger.Error("vote isVoter", "error", err.Error())
		return err
	}

	// get last round vote record.
	lastVoteRecord, ok := os.voteRecords[os.curRound-1]
	if !ok {
		os.logger.Debug("no last round data, reports just with current round commitment")
	}

	// if node is no longer a validator, and it doesn't have last round data, skip reporting.
	if !isVoter && !ok {
		os.logger.Debug("skip data reporting since client is no longer a voter")
		if metrics.Enabled {
			metrics.GetOrRegisterGauge(monitor.IsVoterMetric, nil).Update(0)
		}
		return nil
	}

	// check with the vote buffer from the last penalty event.
	if os.memories.outlierRecord != nil && os.curRoundHeight-os.memories.outlierRecord.LastPenalizedAtBlock <= os.conf.VoteBuffer {
		left := os.conf.VoteBuffer - (os.curRoundHeight - os.memories.outlierRecord.LastPenalizedAtBlock)
		os.logger.Warn("due to the outlier penalty, we postpone your next vote from slashing", "next vote block", left)
		os.logger.Warn("your last outlier report was", "report", os.memories.outlierRecord)
		os.logger.Warn("during this period, you can: 1. check your data source infra; 2. restart your oracle-client; 3. contact Autonity team for help;")
		return nil
	}

	if isVoter {
		// a voter need to assemble current round data to report it.
		curVoteRecord, err := os.buildVoteRecord(os.curRound)
		if err != nil {
			// skipping round vote does not introduce reveal failure.
			os.logger.Info("skip current round vote", "height", os.curRoundHeight, "err", err.Error())
			return err
		}
		if metrics.Enabled {
			metrics.GetOrRegisterGauge(monitor.IsVoterMetric, nil).Update(1)
		}
		// report with last round data and with current round commitment hash.
		return os.reportWithCommitment(curVoteRecord, lastVoteRecord)
	}

	// edge case, voter is no longer a committee member, it has to reveal the last round data that it committed to.
	if lastVoteRecord != nil {
		return os.reportWithoutCommitment(lastVoteRecord)
	}
	return nil
}

// reportWithCommitment reports the commitment of current round, and with last round data if the last round data is available.
// if the input last round data is nil, we just need to report the commitment of current round without last round data.
func (os *Server) reportWithCommitment(curVoteRecord, lastVote *types.VoteRecord) error {
	// prepare the transaction which carry current round's commitment, and last round's data.
	tx, err := os.doReport(curVoteRecord.CommitmentHash, lastVote)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}

	os.logger.Info("reported last round data and with current round commitment", "TX hash", tx.Hash(),
		"Nonce", tx.Nonce(), "Cost", tx.Cost())

	// alert in case of balance reach the warning value.
	balance, err := os.client.BalanceAt(context.Background(), os.conf.Key.Address, nil)
	if err != nil {
		os.logger.Error("cannot get account balance", "error", err.Error())
		return err
	}

	if metrics.Enabled {
		metrics.GetOrRegisterGauge(monitor.BalanceMetric, nil).Update(balance.Int64())
	}

	os.logger.Info("oracle server account", "address", os.conf.Key.Address, "remaining balance", balance.String())
	if balance.Cmp(alertBalance) <= 0 {
		os.logger.Warn("oracle account has too less balance left for data reporting", "balance", balance.String())
	}

	// round data was successfully assembled, save current round data.
	curVoteRecord.TxHash = tx.Hash()
	curVoteRecord.TxNonce = tx.Nonce()
	curVoteRecord.TxCost = tx.Cost()
	os.voteRecords[os.curRound] = curVoteRecord
	if err = os.memories.flushRecord(os.voteRecords); err != nil {
		os.logger.Warn("failed to flush vote record to persistence", "error", err.Error())
		os.logger.Warn("IMPORTANT: please check your profile data dir, the server need to flush vote record into it.")
	}
	return nil
}

// report with last round data but without current round commitment, voter is leaving from the committee.
func (os *Server) reportWithoutCommitment(lastVoteRecord *types.VoteRecord) error {

	// report with no commitment of current round as voter is leaving from the committee.
	tx, err := os.doReport(common.Hash{}, lastVoteRecord)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}
	os.logger.Info("reported last round data and without current round commitment", "TX hash", tx.Hash(), "Nonce", tx.Nonce())

	// save current vote record even though there is no commitment as the voter is leaving the committee.
	curVoteRecord := &types.VoteRecord{
		RoundHeight: os.curRoundHeight,
		RoundID:     os.curRound,
		VotePeriod:  os.votePeriod,
		TxCost:      tx.Cost(),
		TxNonce:     tx.Nonce(),
		TxHash:      tx.Hash(),
	}
	os.voteRecords[os.curRound] = curVoteRecord
	if err = os.memories.flushRecord(os.voteRecords); err != nil {
		os.logger.Warn("failed to flush vote record to persistence", "error", err.Error())
		os.logger.Warn("IMPORTANT: please check your profile data dir, the server need to flush vote record into it.")
	}
	return nil
}

// resetSamplingSymbols reset the latest sampling symbol set with the protocol symbol set.
func (os *Server) resetSamplingSymbols(protocolSymbols []string) {
	os.samplingSymbols = protocolSymbols

	// check if we need to add bridger symbols on demand.
	bridged := false
	for _, s := range protocolSymbols {
		if bridger, ok := BridgedSymbols[s]; ok {
			bridged = true
			os.samplingSymbols = append(os.samplingSymbols, bridger)
		}
	}

	if bridged {
		os.samplingSymbols = append(os.samplingSymbols, USDCUSD)
	}
}

// addNewSymbols adds new symbols to the local symbol set for data fetching, duplicated one is not added.
func (os *Server) addNewSymbols(newSymbols []string) {
	var symbolsMap = make(map[string]struct{})
	for _, s := range os.samplingSymbols {
		symbolsMap[s] = struct{}{}
	}

	// check if we need to add bridger symbols on demand.
	bridged := false
	for _, newS := range newSymbols {
		if _, ok := symbolsMap[newS]; !ok {
			os.samplingSymbols = append(os.samplingSymbols, newS)
			// if the new symbol requires a bridger symbol, add it too.
			if bridger, ok := BridgedSymbols[newS]; ok {
				bridged = true
				if _, ok := symbolsMap[bridger]; !ok {
					os.samplingSymbols = append(os.samplingSymbols, bridger)
				}
			}
		}
	}

	if _, ok := symbolsMap[USDCUSD]; !ok && bridged {
		os.samplingSymbols = append(os.samplingSymbols, USDCUSD)
	}
}

func (os *Server) resolveGasTipCap() *big.Int {
	configured := new(big.Int).SetUint64(os.conf.GasTipCap)
	suggested, err := os.client.SuggestGasTipCap(context.Background())
	if err != nil {
		os.logger.Warn("cannot get fee history, using configured gas tip cap", "error", err.Error())
		return configured
	}

	// take the max one to let the report get mine with higher priority.
	if suggested.Cmp(configured) > 0 {
		return suggested
	}
	return configured
}

func (os *Server) doReport(curRoundCommitmentHash common.Hash, lastVoteRecord *types.VoteRecord) (*tp.Transaction, error) {
	chainID, err := os.client.ChainID(context.Background())
	if err != nil {
		os.logger.Error("get chain id", "error", err.Error())
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(os.conf.Key.PrivateKey, chainID)
	if err != nil {
		os.logger.Error("new keyed transactor with chain ID", "error", err)
		return nil, err
	}

	gasTipCap := os.resolveGasTipCap()
	// Get base fee for pending block
	header, err := os.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		os.logger.Error("get header by number", "error", err.Error())
		return nil, err
	}

	// Calculate MaxFeePerGas (baseFee * 2 + gasTipCap)
	maxFeePerGas := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, gasTipCap)

	auth.Value = big.NewInt(0)
	auth.GasTipCap = gasTipCap
	auth.GasFeeCap = maxFeePerGas
	auth.GasLimit = uint64(3000000)

	// if there is no last round data, it could be the client was omission faulty at last round, then we just submit the
	// commitment hash of current round. If we cannot recover the last round vote record from persistence layer, then
	// below vote without data could lead to reveal failure still.
	if lastVoteRecord == nil {
		var reports []contract.IOracleReport
		return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), reports, invalidSalt, config.Version)
	}

	// there is last round data, report with current round commitment, and the last round reports and salt to be revealed.
	return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), lastVoteRecord.Reports, lastVoteRecord.Salt, config.Version)
}

func (os *Server) buildVoteRecord(round uint64) (*types.VoteRecord, error) {
	if len(os.protocolSymbols) == 0 {
		return nil, types.ErrNoSymbolsObserved
	}

	prices, err := os.aggregateProtocolSymbolPrices()
	if err != nil {
		return nil, err
	}

	// assemble round data with reports, salt and commitment hash.
	voteRecord, err := os.assembleVote(round, os.protocolSymbols, prices)
	if err != nil {
		os.logger.Error("failed to assemble round report data", "error", err.Error())
		return nil, err
	}
	os.logger.Info("assembled round report data", "current round", round, "prices", voteRecord)
	return voteRecord, nil
}

func (os *Server) aggregateProtocolSymbolPrices() (types.PriceBySymbol, error) {
	prices := make(types.PriceBySymbol)

	// if we need a bridger pair USDC-USD to convert ATN-USD or NTN-USD from ATN-USDC or NTN-USDC,
	// then we need to aggregate USDC-USD data point first.
	var usdcPrice *types.Price
	var err error
	if slices.Contains(os.protocolSymbols, ATNUSD) || slices.Contains(os.protocolSymbols, NTNUSD) {
		usdcPrice, err = os.aggregatePrice(USDCUSD, os.curSampleTS)
		if err != nil {
			os.logger.Error("aggregate USDC-USD price", "error", err.Error())
		}
	}

	for _, s := range os.protocolSymbols {
		// aggregate bridged symbols
		if s == ATNUSD || s == NTNUSD {
			if usdcPrice == nil {
				continue
			}

			p, e := os.aggBridgedPrice(s, os.curSampleTS, usdcPrice)
			if e != nil {
				os.logger.Error("aggregate bridged price", "error", e.Error(), "symbol", s)
				continue
			}
			prices[s] = *p
			continue
		}

		// aggregate none bridged symbols
		p, e := os.aggregatePrice(s, os.curSampleTS)
		if e != nil {
			os.logger.Debug("no data for aggregation", "reason", e.Error(), "symbol", s)
			continue
		}
		prices[s] = *p
	}

	// edge case: if NTN-ATN price was not computable from inside plugin,
	// try to compute it from NTNprice and ATNprice across from different plugins.
	if _, ok := prices[common2.NTNATNSymbol]; !ok {
		ntnPrice, ntnExist := prices[NTNUSD]
		atnPrice, atnExist := prices[ATNUSD]
		if ntnExist && atnExist {
			ntnATNPrice, err := common2.ComputeDerivedPrice(ntnPrice.Price.String(), atnPrice.Price.String()) //nolint
			if err == nil {
				p, err := decimal.NewFromString(ntnATNPrice.Price) // nolint
				if err == nil {
					prices[common2.NTNATNSymbol] = types.Price{
						Timestamp:  time.Now().Unix(),
						Price:      p,
						Symbol:     common2.NTNATNSymbol,
						Confidence: ntnPrice.Confidence,
					}
				} else {
					os.logger.Error("cannot parse NTN-ATN price in decimal", "error", err.Error())
				}
			} else {
				os.logger.Error("failed to compute NTN-ATN price", "error", err.Error())
			}
		}
	}

	return prices, nil
}

// assemble the final reports, salt and commitment hash.
func (os *Server) assembleVote(round uint64, symbols []string, prices types.PriceBySymbol) (*types.VoteRecord, error) {
	var voteRecord = &types.VoteRecord{
		RoundHeight: os.curRoundHeight,
		RoundID:     round,
		VotePeriod:  os.votePeriod,
		Symbols:     symbols,
		Prices:      prices,
	}

	var missingData bool
	var reports []contract.IOracleReport
	for _, s := range symbols {
		if pr, ok := prices[s]; ok {
			// This is an edge case, which means there is no liquidity in the market for this symbol.
			price := pr.Price.Mul(os.pricePrecision).BigInt()
			if price.Cmp(invalidPrice) == 0 {
				os.logger.Info("please check your data source, zero data point measured from market", "symbol", s)
				missingData = true
			}
			reports = append(reports, contract.IOracleReport{
				Price:      price,
				Confidence: pr.Confidence,
			})
		} else {
			// logging the missing of data points for symbols
			missingData = true
			os.logger.Info("please check your data source, missing data point for symbol", "symbol", s)
		}
	}

	// we won't assemble the round data if any data point is missing.
	if missingData {
		return nil, types.ErrMissingDataPoint
	}

	salt, err := rand.Int(rand.Reader, saltRange)
	if err != nil {
		os.logger.Error("generate rand salt", "error", err.Error())
		return nil, err
	}

	commitmentHash, err := os.commitmentHashComputer.CommitmentHash(reports, salt, os.conf.Key.Address)
	if err != nil {
		os.logger.Error("failed to compute commitment hash", "error", err.Error())
		return nil, err
	}

	voteRecord.Reports = reports
	voteRecord.Salt = salt
	voteRecord.CommitmentHash = commitmentHash
	return voteRecord, nil
}

func (os *Server) handleNewSymbolsEvent(symbols []string) {
	// just add symbols to oracle service's symbol pool, thus the oracle service can start to prepare the data.
	os.addNewSymbols(symbols)
}

// aggBridgedPrice aggregates ATN-USD or NTN-USD from bridged ATN-USDC or NTN-USDC with USDC-USD price,
// it assumes the input usdcPrice is not nil.
func (os *Server) aggBridgedPrice(srcSymbol string, target int64, usdcPrice *types.Price) (*types.Price, error) {
	var bridgedSymbol string
	if srcSymbol == ATNUSD {
		bridgedSymbol = ATNUSDC
	}

	if srcSymbol == NTNUSD {
		bridgedSymbol = NTNUSDC
	}

	p, err := os.aggregatePrice(bridgedSymbol, target)
	if err != nil {
		os.logger.Error("aggregate bridged price", "error", err.Error(), "symbol", bridgedSymbol)
		return nil, err
	}

	// reset the symbol with source symbol,
	// and update price with: ATN-USD=ATN-USDC*USDC-USD / NTN-USD=NTN-USDC*USDC-USD
	// the confidence of ATN-USD and NTN-USD are inherit from ATN-USDC and NTN-USDC.
	p.Symbol = srcSymbol
	p.Price = p.Price.Mul(usdcPrice.Price)
	return p, nil
}

// aggregatePrice takes the symbol's aggregated data points from all the supported plugins, if there are multiple
// markets' datapoint, it will do a final VWAP aggregation to form the final reporting value.
func (os *Server) aggregatePrice(s string, target int64) (*types.Price, error) {
	prices, volumes := os.pluginManager.SelectSamples(s, target)
	if len(prices) == 0 {
		copyHistoricPrice, err := os.queryHistoricRoundPrice(s)
		if err != nil {
			return nil, err
		}

		return confidenceAdjustedPrice(&copyHistoricPrice, target)
	}

	// compute confidence of the symbol from the num of plugins' samples of it.
	confidence := computeConfidence(s, len(prices), os.conf.ConfidenceStrategy)
	price := &types.Price{
		Timestamp:  target,
		Price:      prices[0],
		Volume:     volumes[0],
		Symbol:     s,
		Confidence: confidence,
	}

	_, isForex := common2.ForexCurrencies[s]

	// we have multiple markets' data for this forex symbol, update the price with median value.
	if len(prices) > 1 && isForex {
		p, err := helpers.Median(prices)
		if err != nil {
			return nil, err
		}
		price.Price = p
		price.Volume = types.DefaultVolume
		return price, nil
	}

	// we have multiple markets' data for this crypto symbol, update the price with VWAP.
	if len(prices) > 1 && !isForex {
		p, vol, err := helpers.VWAP(prices, volumes)
		if err != nil {
			return nil, err
		}
		price.Price = p
		price.Volume = vol
	}

	return price, nil
}

// queryHistoricRoundPrice queries the last available price for a given symbol from the historic rounds.
func (os *Server) queryHistoricRoundPrice(symbol string) (types.Price, error) {

	if len(os.voteRecords) == 0 {
		return types.Price{}, types.ErrNoDataRound
	}

	numOfRounds := len(os.voteRecords)
	// Iterate from the current round backward
	for i := 0; i < numOfRounds; i++ {
		roundID := os.curRound - uint64(i) - 1 //nolint
		// Get the round data for the current round ID
		voteRecord, exists := os.voteRecords[roundID]
		if !exists {
			continue
		}

		if voteRecord == nil {
			continue
		}

		// Check if the symbol exists in the Prices map
		if price, found := voteRecord.Prices[symbol]; found {
			return price, nil
		}
	}

	// If no price was found after checking all rounds, return an error
	return types.Price{}, types.ErrNoDataRound
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
