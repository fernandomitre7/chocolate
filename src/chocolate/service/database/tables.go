package database

import (
	"chocolate/service/shared/logger"
)

func createTables(db *DB) (err error) {
	if err = createTableUsers(db); err != nil {
		return err
	}
	return
}

// createTableUsers tries to create the users table. If it already exists or not,
// The operation only fails in case there's a mismatch in table
// definition of if there's a connection error.
func createTableUsers(db *DB) (err error) {
	const qry = `
		CREATE TABLE IF NOT EXISTS users (
			id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
			username text NOT NULL UNIQUE,
			password text NOT NULL, 
			salt text NOT NULL, 
			confirmed boolean NOT NULL DEFAULT FALSE,
			confirmation_date timestamp with time zone,
			names text,
			first_last_name text,
			second_last_name text,
			created_at timestamp with time zone DEFAULT current_timestamp
		)`

	// Exec executes a query without returning any rows.
	if _, err = db.dbsql.Exec(qry); err != nil {
		logger.Errorf("UsersDB:createTable() Users table creation query failed %s", err.Error())
		return
	}
	// Add Unique Email Index create unique index users_unique_lower_email_idx on users (lower(email));
	const indexQry = `CREATE UNIQUE INDEX users_unique_username_idx on users(lower(username))`
	// Exec executes a query without returning any rows.
	if _, err := db.dbsql.Exec(indexQry); err != nil {
		logger.Errorf("UsersDB:createTable() Users table unique index creation failed %s", err.Error())
		// should we delete the table?
	}

	return
}
