package databases

import "time"

type BadgerQuery interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}) error
	Delete(key string) error
}

type ExtendedQuery interface {
	PrefixScan(prefix string) ([]string, [][]byte, error)
	PrefixScanWithLimit(prefix string, limit int) ([]string, [][]byte, error)
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	BadgerQuery
}

type BadgerBatchQuery interface {
	BatchSet(params []*BadgerSetParams) error
	BatchDelete(params []*BadgerDeleteParams) error
}
