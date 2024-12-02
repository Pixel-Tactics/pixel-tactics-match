package databases

type BadgerQuery interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}) error
	Delete(key string) error
}

type BadgerBatchQuery interface {
	BatchSet(params []*BadgerSetParams) error
	BatchDelete(params []*BadgerDeleteParams) error
}
