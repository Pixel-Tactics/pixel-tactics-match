package databases

type TransactionManager interface {
	NewReadTransaction() BadgerTx
	NewReadWriteTransaction() BadgerTx
}
