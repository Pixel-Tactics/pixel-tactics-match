package databases

import "github.com/dgraph-io/badger/v4"

func NotFoundException() error {
	return badger.ErrKeyNotFound
}
