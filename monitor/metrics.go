package monitor

import "github.com/ethereum/go-ethereum/metrics"

var (
	PluginMetric         = "oracle/plugins"
	RoundMetric          = "oracle/round"
	BalanceMetric        = "oracle/balance"
	IsVoterMetric        = "oracle/isVoter"
	L1ConnectivityMetric = "oracle/l1/errs"
	InvalidVoteMetric    = "oracle/vote/invalid"
	NoRevealVoteMetric   = "oracle/vote/noreveal"
	SuccessfulVoteMetric = "oracle/vote/successful"

	OutlierDistancePercentMetric = "oracle/outlier/distance/percentage"
	OutlierNoSlashTimesMetric    = "oracle/outlier/noslash/times"
	OutlierSlashTimesMetric      = "oracle/outlier/slash/times"
	OutlierPenaltyMetric         = "oracle/outlier/penality"
)

func InitOracleMetrics() {
	if metrics.Enabled {
		metrics.GetOrRegisterGauge(PluginMetric, nil)
		metrics.GetOrRegisterGauge(RoundMetric, nil)
		metrics.GetOrRegisterGauge(BalanceMetric, nil)
		metrics.GetOrRegisterGauge(IsVoterMetric, nil)
		metrics.GetOrRegisterCounter(L1ConnectivityMetric, nil)
		metrics.GetOrRegisterCounter(InvalidVoteMetric, nil)
		metrics.GetOrRegisterCounter(NoRevealVoteMetric, nil)
		metrics.GetOrRegisterCounter(SuccessfulVoteMetric, nil)

		// create metrics for outlier penalty events in advance.
		metrics.GetOrRegisterGauge(OutlierDistancePercentMetric, nil)
		metrics.GetOrRegisterCounter(OutlierNoSlashTimesMetric, nil)
		metrics.GetOrRegisterCounter(OutlierSlashTimesMetric, nil)
		metrics.GetOrRegisterGaugeFloat64(OutlierPenaltyMetric, nil)
	}
}
