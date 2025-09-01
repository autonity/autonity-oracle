package server

import (
	"autonity-oracle/config"
	"autonity-oracle/contract_binder/contract"
	"autonity-oracle/monitor"
	types2 "autonity-oracle/types"
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/metrics"
)

func (os *Server) handlePenaltyEvent(penalizeEvent *oracle.OraclePenalized) error {
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
func (os *Server) unpackLog(out interface{}, event string, log types.Log) error {
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
		Addresses: []common.Address{types2.OracleContractAddress},
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
	ev := new(oracle.OraclePenalized)
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
		return types2.ErrPeerOnSync
	}

	// sync protocol symbols before vote.
	if err := os.syncProtocolSymbols(); err != nil {
		return err
	}

	// as outlier slashing event can come right after round event in the same block.
	// if node is on outlier slashing, skip round vote to avoid the outlier slashing again.
	if os.checkOutlierSlashing() {
		return types2.ErrOnOutlierSlashing
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
func (os *Server) reportWithCommitment(curVoteRecord, lastVote *types2.VoteRecord) error {
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
func (os *Server) reportWithoutCommitment(lastVoteRecord *types2.VoteRecord) error {

	// report with no commitment of current round as voter is leaving from the committee.
	tx, err := os.doReport(common.Hash{}, lastVoteRecord)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}
	os.logger.Info("reported last round data and without current round commitment", "TX hash", tx.Hash(), "Nonce", tx.Nonce())

	// save current vote record even though there is no commitment as the voter is leaving the committee.
	curVoteRecord := &types2.VoteRecord{
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

func (os *Server) doReport(curRoundCommitmentHash common.Hash, lastVoteRecord *types2.VoteRecord) (*types.Transaction, error) {
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
		var reports []oracle.IOracleReport
		return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), reports, invalidSalt, config.Version)
	}

	// there is last round data, report with current round commitment, and the last round reports and salt to be revealed.
	return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), lastVoteRecord.Reports, lastVoteRecord.Salt, config.Version)
}

func (os *Server) buildVoteRecord(round uint64) (*types2.VoteRecord, error) {
	if len(os.protocolSymbols) == 0 {
		return nil, types2.ErrNoSymbolsObserved
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

// assemble the final reports, salt and commitment hash.
func (os *Server) assembleVote(round uint64, symbols []string, prices types2.PriceBySymbol) (*types2.VoteRecord, error) {
	var voteRecord = &types2.VoteRecord{
		RoundHeight: os.curRoundHeight,
		RoundID:     round,
		VotePeriod:  os.votePeriod,
		Symbols:     symbols,
		Prices:      prices,
	}

	var missingData bool
	var reports []oracle.IOracleReport
	for _, s := range symbols {
		if pr, ok := prices[s]; ok {
			// This is an edge case, which means there is no liquidity in the market for this symbol.
			price := pr.Price.Mul(os.pricePrecision).BigInt()
			if price.Cmp(invalidPrice) == 0 {
				os.logger.Info("please check your data source, zero data point measured from market", "symbol", s)
				missingData = true
			}
			reports = append(reports, oracle.IOracleReport{
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
		return nil, types2.ErrMissingDataPoint
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

// trackVoteState works in a pull mode to track if the vote was mined by L1 although there is already a push mode
// which subscribe the events from oracle contract. We cannot sure that L1 node is already on service, for example,
// some operation, synchronization, resetting, etc...
func (os *Server) trackVoteState() {

	var update bool
	endRound := os.curRound - MaxBufferedRounds
	if os.curRound <= MaxBufferedRounds {
		endRound = 0
	}
	for r := os.curRound; r > endRound; r-- {
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
	endRound := os.curRound - MaxBufferedRounds
	if os.curRound <= MaxBufferedRounds {
		endRound = 0
	}
	// iterate from the most recent round.
	for r := os.curRound; r > endRound; r-- {
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
