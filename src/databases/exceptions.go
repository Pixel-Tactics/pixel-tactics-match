package databases

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
)

var ErrNilTransaction = errors.New("nil transaction")

func NotFoundException() error {
	return badger.ErrKeyNotFound
}
