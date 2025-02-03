package databases

func GetQuery(tx BadgerTx, def BadgerQuery) BadgerQuery {
	if tx != nil {
		return tx
	}
	return def
}
