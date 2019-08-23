package ck

import (
	"sync"
)

type memCKAPI struct {
	sync.RWMutex

	tables map[string][]uint64
}

// NewMemCKAPI returns a implementation by mem
func NewMemCKAPI() API {
	return &memCKAPI{
		tables: make(map[string][]uint64),
	}
}

func (api *memCKAPI) Remove(table string, partitions ...uint64) error {
	api.Lock()
	defer api.Unlock()

	if ps, ok := api.tables[table]; ok {
		var values []uint64
		for _, p := range ps {
			found := false
			for _, rm := range partitions {
				if p == rm {
					found = true
				}
			}

			if !found {
				values = append(values, p)
			}
		}

		api.tables[table] = values
	}

	return nil
}

func (api *memCKAPI) Create(table string, partitions ...uint64) error {
	api.Lock()
	defer api.Unlock()

	if ps, ok := api.tables[table]; ok {
		var values []uint64
		for _, p := range ps {
			found := false
			for _, added := range partitions {
				if p == added {
					found = true
				}
			}

			if !found {
				values = append(values, p)
			}
		}

		ps = append(ps, values...)
		api.tables[table] = ps
	} else {
		api.tables[table] = partitions
	}

	return nil
}
