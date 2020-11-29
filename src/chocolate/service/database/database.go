// Package database wraps `lib/pq` providing the basic methods for creating an entrypoint for our database.
package database

import (
	"database/sql"
	"fmt"
	"strings"

	// Using the blank identifier in order to solely
	// provide the side-effects of the package.
	// Eseentially the side effect is calling the `init()`
	// method of `lib/pq`:
	//	func init () {  sql.Register("postgres", &Driver{} }
	// which you can see at `github.com/lib/pq/conn.go`
	"errors"

	"chocolate/service/shared/config"
	"chocolate/service/shared/logger"

	"github.com/lib/pq"
)

const (
	connectionErrorClass = "08"
	internalErrorClass   = "XX"
)

// DB holds the connection pool to the database - created by a configuration object (`SQLConfig`).
type DB struct {
	// dbsql holds a sql.DB pointer that represents a pool of zero or more
	// underlying connections - safe for concurrent use by multiple
	// goroutines -, with freeing/creation of new connections all managed
	// by `sql/database` package.
	dbsql *sql.DB
	// The DB configuration
	cfg config.SQLConfig
}

// Table is the interface needed by the API Models to be able to create themselves,
// This means creating the Tables and Indexes needed for the specific Model
type Table interface {
	Create(db *DB) (dberr *Error)
	Name() string
}

// GetInstance returns the actual sql.DB
func (db DB) GetInstance() *sql.DB {
	return db.dbsql
}

// New returns a SQL DB with the sql.DB set with the postgres
// DB connection string in the configuration
func New(cfg config.SQLConfig) (db *DB, err error) {
	logger.Infof("Starting DB at host: %s port: %s...", cfg.Host, cfg.Port)
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" ||
		cfg.Password == "" || cfg.Database == "" {
		err = errors.New("All db configuration fields must be set")
		return
	}

	// The first argument corresponds to the driver name that the driver
	// (in this case, `lib/pq`) used to register itself in `database/sql`.
	// The next argument specifies the parameters to be used in the connection.
	// Details about this string can be seen at https://godoc.org/github.com/lib/pq
	connString := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.User, cfg.Password, cfg.Database, cfg.Host, cfg.Port)
	logger.Debugf("DB Connection string: %s", connString)

	dbsql, err := sql.Open("postgres", connString)
	if err != nil {
		err = fmt.Errorf("Couldn't open connection to postgre database: %s", err.Error())
		return
	}
	logger.Infof("Pinging DB...")
	// Ping verifies if the connection to the database is alive or if a
	// new connection can be made.
	if err = dbsql.Ping(); err != nil {
		err = fmt.Errorf("Couldn't ping postgres database: %s", err.Error())
		return
	}

	db = &DB{dbsql, cfg}

	/* if err = createTables(db); err != nil {
		logger.Errorf("Couldnt create tables: %s", err.Error())
		return
	} */
	return
}

// Close performs the release of any resources that
// `sql/database` DB pool created. This is usually meant
// to be used in the exitting of a program or `panic`ing.
func (db *DB) Close() (err error) {
	if db.dbsql == nil {
		return
	}

	if err = db.dbsql.Close(); err != nil {
		err = fmt.Errorf("Errored closing database connection: %s", err.Error())
	}

	return
}

func installExtensions(db *DB) *Error {
	qry := `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`
	if _, err := db.dbsql.Exec(qry); err != nil {
		logger.Errorf("database:installExtensions() Failed to install extension err: %s", err.Error())
		return db.FormError(err, qry, "extensions")
	}
	return nil
}

// Init will hold any type of initialization logic needed by the DB,
// Right now it only creates the needed Model's tables
func (db *DB) Init(tables []Table) *Error {
	// check that extensions are properly installed

	extQry := `SELECT uuid_generate_v4()`

	if _, err := db.dbsql.Exec(extQry); err != nil {
		logger.Errorf("database:Init() checking of extensions failed %s", err.Error())
		dberr := db.FormError(err, extQry, "")
		if dberr.Code == ErrorMissingExtensions {
			logger.Infof("database:Init() missing uuid extensions, installing now")
			if dberr = installExtensions(db); dberr != nil {
				return dberr
			}
		} else {
			return dberr
		}
	}

	for _, table := range tables {

		logger.Infof("Creating table %q", table.Name())
		if dberr := table.Create(db); dberr != nil {
			// NOTE: what happens if one table is created but another isn't?
			// I guess is OK, we need the tables and they all should have the "IF NOT EXISTS" clause
			logger.Errorf("Error creating table %q. err = %v", table.Name(), dberr.Error())
			return dberr
		}
	}

	return nil
}

// FormError returns the pq(postgres) error wrapped in a Error
func (db *DB) FormError(err error, query, table string) (dberr *Error) {
	logger.Debugf("DB.FormError() err = %v, table = %q, query = %q", err, table, query)
	if err == sql.ErrNoRows {
		dberr = NewError(ErrorNoRows, "No rows found", query, table, err)
		return
	}
	if pqerr, ok := err.(*pq.Error); ok {
		code := string(pqerr.Code)
		logger.Debugf("pq error %s:", code)

		if strings.HasPrefix(code, connectionErrorClass) {
			dberr = NewError(ErrorInternal, "Connection Error", query, table, pqerr)
		} else if strings.HasPrefix(code, internalErrorClass) {
			dberr = NewError(ErrorInternal, "Internal DB Error", query, table, pqerr)
		} else if code == "42883" {
			// 42883 = "undefined_function"
			dberr = NewError(ErrorMissingExtensions, "Missing Postgres extension", query, table, pqerr)
		} else if code == "23505" {
			// unique constrain violation
			dberr = NewError(ErrorAlreadyExists, "Already Exists, unique constrain violation", query, table, pqerr)
		} else if code == "02000" {
			// "02000": "no_data"
			dberr = NewError(ErrorNoData, "No Data", query, table, pqerr)
		} else {
			dberr = NewError(ErrorExecute, "Error executing query", query, table, pqerr)
		}

		return

	}
	logger.Errorf("This is not a PQ or SQL error: %s", err.Error())
	dberr = NewError(ErrorGeneric, "", query, table, err)

	return
}
