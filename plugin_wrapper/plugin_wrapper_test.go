package pluginwrapper

import (
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPluginWrapper(t *testing.T) {
	t.Run("test finding nearest data sample", func(t *testing.T) {
		p := PluginWrapper{
			samples:          make(map[string]map[int64]types.Price),
			latestTimestamps: make(map[string]int64),
		}

		now := time.Now().Unix()
		for ts := now; ts < now+60; ts++ {
			if now+29 <= ts && ts < now+35 {
				continue
			}

			var prices []types.Price
			prices = append(prices, types.Price{
				Timestamp: ts,
				Symbol:    "NTNGBP",
				Price:     decimal.RequireFromString("1.1"),
			})
			p.AddSample(prices, ts)
		}

		target := now
		price, err := p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now, price.Timestamp)

		// upper bound
		target = now + 100
		price, err = p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now+59, price.Timestamp)

		// lower bound
		target = now - 1
		price, err = p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now, price.Timestamp)

		// middle
		target = now + 29
		price, err = p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now+28, price.Timestamp)

		// middle
		target = now + 33
		price, err = p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now+35, price.Timestamp)

		// middle
		target = now + 34
		price, err = p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now+35, price.Timestamp)

		// middle
		target = now + 35
		price, err = p.GetSample("NTNGBP", target)
		require.NoError(t, err)
		require.Equal(t, now+35, price.Timestamp)

		// test gc, at least 1 sample is kept in the cache.
		p.GCSamples()
		require.Equal(t, 1, len(p.samples))
	})
}
