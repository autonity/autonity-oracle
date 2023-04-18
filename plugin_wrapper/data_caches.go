package pluginwrapper

import (
	"autonity-oracle/types"
	"fmt"
	"sync"
)

type DataCache struct {
	name         string
	lock         sync.RWMutex
	priceSamples map[string]map[int64]types.Price
}

func NewDataCache(name string) *DataCache {
	return &DataCache{
		name:         name,
		priceSamples: make(map[string]map[int64]types.Price),
	}
}

func (t *DataCache) AddSample(prices []types.Price, ts int64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, p := range prices {
		tsMap, ok := t.priceSamples[p.Symbol]
		if !ok {
			tsMap = make(map[int64]types.Price)
			tsMap[ts] = p
			t.priceSamples[p.Symbol] = tsMap
			return
		}
		tsMap[ts] = p
	}
}

func (t *DataCache) GetSample(symbol string, ts int64) (types.Price, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	tsMap, ok := t.priceSamples[symbol]
	if !ok {
		return types.Price{}, fmt.Errorf("symbol not find")
	}

	p, ok := tsMap[ts]
	if !ok {
		return types.Price{}, types.ErrNoAvailablePrice
	}

	return p, nil
}

func (t *DataCache) GCSamples() {
	for k := range t.priceSamples {
		delete(t.priceSamples, k)
	}
}

type DataCacheSet struct {
	lock        sync.RWMutex
	providerSet map[string]*DataCache
}

func NewDataCacheSet() *DataCacheSet {
	return &DataCacheSet{
		providerSet: make(map[string]*DataCache),
	}
}

func (tp *DataCacheSet) AddDataCache(provider string) *DataCache {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	tp.providerSet[provider] = NewDataCache(provider)
	return tp.providerSet[provider]
}

func (tp *DataCacheSet) GetDataCache(provider string) *DataCache {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	return tp.providerSet[provider]
}

func (tp *DataCacheSet) DeleteDataCache(provider string) {
	tp.lock.Lock()
	defer tp.lock.Unlock()
	delete(tp.providerSet, provider)
}
