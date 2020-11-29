package models

import (
	"chocolate/service/database"
	"chocolate/service/models/users"
)

// APIObject Interface all the service Models should implement
type APIObject interface {
	JSON() ([]byte, error)
	Valid() error
	Decode(data []byte) (err error)
}

// GetDBTables returns the tables needed to support the Models
// This is where the DB and Models are connected (the dependency is created)
// NOT EVERY model needs to have a table, only whatever needs to be persisted
func GetDBTables() []database.Table {
	return []database.Table{
		users.GetTable(),
	}
}
