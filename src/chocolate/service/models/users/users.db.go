package users

import (
	"database/sql"
	"fmt"
	"time"

	"chocolate/service/database"
	"chocolate/service/shared/logger"
)

const (
	qryAll     = `id, username, password, salt, confirmed, confirmed_at, created_at`
	qryAllSafe = `id, username, confirmed, confirmed_at, created_at`

	qryCreateTable = `CREATE TABLE IF NOT EXISTS users (
		id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
		username text NOT NULL UNIQUE,
		password text NOT NULL, 
		salt text NOT NULL, 
		confirmed boolean NOT NULL DEFAULT FALSE,
		confirmation_date timestamp with time zone,
		created_at timestamp with time zone DEFAULT current_timestamp
	)`
)

type userTable struct{}

func (t userTable) Create(db *database.DB) *database.Error {
	logger.Debug("models/userTable:Create() exec qry")
	// Exec executes a query without returning any rows.
	if _, err := db.GetInstance().Exec(qryCreateTable); err != nil {
		logger.Errorf("models/users:createTable() Users table creation query failed %s", err.Error())
		return db.FormError(err, qryCreateTable, "users")
	}
	// Add Unique Email Index create unique index users_unique_lower_email_idx on users (lower(email));
	const indexQry = `CREATE UNIQUE INDEX IF NOT EXISTS users_unique_username_idx on users(lower(username))`
	logger.Debug("models/userTable:Create() exec index qry")
	// Exec executes a query without returning any rows.
	if _, err := db.GetInstance().Exec(indexQry); err != nil {
		logger.Errorf("models/users:createTable() Users table unique index creation failed %s", err.Error())
		// should we delete the table?
		return db.FormError(err, indexQry, "users")
	}

	return nil
}

func (t userTable) Name() string {
	return "users"
}

// GetTable returns the Model Table struct which is used to create the necessary db tables for the Model
func GetTable() database.Table {
	return userTable{}
}

// GetBy gets a User by field and value
func GetBy(db *database.DB, field string, value interface{}, reqID string) (u User, dberr *database.Error) {

	qry := fmt.Sprintf(`SELECT %s FROM users WHERE %s = $1`, qryAll, field)

	// `QueryRow` is a single-row query that, unlike `Query()`, doesn't hold a connection.
	// Errors from `QueryRow` are forwarded to `Scan` where we can get errors from both.
	// Here we perform such query for inserting because we want to grab right from the Database the entry that was inserted
	// (plus the fields that the database generated).
	// If we were just getting a value, we could also check if the query
	// was successfull but returned 0 rows with `if err == sql.ErrNoRows`.
	var createdAt time.Time
	u = User{}
	row := db.GetInstance().QueryRow(qry, value)
	err := scanAll(row, &u)

	if err != nil {
		logger.Errorf("%v:User:GetBy() Couldn't get user by field = '%s', value = '%v': %s", reqID, field, value, err.Error())
		dberr = db.FormError(err, qry, "users")
		return
	}

	u.CreatedAt = createdAt.Unix()
	return
}

// GetByID gets a User by ID
func GetByID(db *database.DB, userID, reqID string) (u User, dberr *database.Error) {

	qry := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, qryAllSafe)

	// `QueryRow` is a single-row query that, unlike `Query()`, doesn't hold a connection.
	// Errors from `QueryRow` are forwarded to `Scan` where we can get errors from both.
	// Here we perform such query for inserting because we want to grab right from the Database the entry that was inserted
	// (plus the fields that the database generated).
	// If we were just getting a value, we could also check if the query
	// was successfull but returned 0 rows with `if err == sql.ErrNoRows`.
	var createdAt time.Time
	u = User{}
	row := db.GetInstance().QueryRow(qry, userID)
	err := scanAllSafe(row, &u)

	if err != nil {
		logger.Errorf("%v:User:GetByID() Couldn't get user(%s): %s", reqID, userID, err.Error())
		dberr = db.FormError(err, qry, "users")
		return
	}

	u.CreatedAt = createdAt.Unix()
	return
}

// GetList retrieves the list of Users
func GetList(db *database.DB, reqID string) (users Users, dberr *database.Error) {

	qry := fmt.Sprintf(`SELECT %s FROM users`, qryAllSafe)

	rows, err := db.GetInstance().Query(qry)
	if err != nil {
		logger.Errorf("%s:Error Getting list of users: %v", reqID, err)
		dberr = db.FormError(err, qry, "users")
		return
	}

	defer rows.Close()
	for rows.Next() {
		u := User{}
		var createdAt, confirmedAt time.Time
		if err = rows.Scan(&u.ID, &u.Username, &u.Confirmed, &confirmedAt, &createdAt); err != nil {
			logger.Errorf("%s:Error Scanning Row of users: %v", reqID, err)
			dberr = db.FormError(err, qry, "users")
			return
		}
		u.CreatedAt = createdAt.Unix()
		if !confirmedAt.IsZero() {
			u.ConfirmedAt = confirmedAt.Unix()
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		logger.Errorf("%s:Error Scanning in Row of users: %v", reqID, err)
		dberr = db.FormError(err, qry, "users")
		return
	}

	return
}

// Insert creates a User record in DB
func (u *User) Insert(db *database.DB, reqID string) (dberr *database.Error) {

	qry := `INSERT INTO users(username, password, salt, confirmed, confirmed_at) 
			VALUES($1, $2, $3, $4, $5) RETURNING id, created_at`

	// `QueryRow` is a single-row query that, unlike `Query()`, doesn't hold a connection.
	// Errors from `QueryRow` are forwarded to `Scan` where we can get errors from both.
	// Here we perform such query for inserting because we want to grab right from the Database the entry that was inserted
	// (plus the fields that the database generated).
	// If we were just getting a value, we could also check if the query
	// was successfull but returned 0 rows with `if err == sql.ErrNoRows`.
	var createdAt, confirmedAt time.Time
	if u.ConfirmedAt > 0 {
		confirmedAt = time.Unix(u.ConfirmedAt, 0)
	}
	err := db.GetInstance().QueryRow(qry, u.Username, u.Password, u.Salt, u.Confirmed, confirmedAt).Scan(&u.ID, &createdAt)

	if err != nil {
		logger.Errorf("%v:User:Insert() Couldn't insert new user: %s", reqID, err.Error())
		dberr = db.FormError(err, qry, "users")
		return
	}

	u.CreatedAt = createdAt.Unix()
	u.Password = ""
	u.PasswordConfirm = ""
	u.Salt = ""
	return
}

// Update updates current user fields
func (u *User) Update(db *database.DB, reqID string) (dberr *database.Error) {

	qry := `UPDATE users SET confirmed = $2,  confirmed_at = $3 WHERE id = $1 
			RETURNING id, username, confirmed, confirmed_at, created_at`
	if u.ID == "" {
		dberr = database.NewError(database.ErrorModelInvalid, "Missing ID value", qry, "users", nil)
		return
	}
	// `QueryRow` is a single-row query that, unlike `Query()`, doesn't hold a connection.
	// Errors from `QueryRow` are forwarded to `Scan` where we can get errors from both.
	// Here we perform such query for inserting because we want to grab right from the Database the entry that was inserted
	// (plus the fields that the database generated).
	// If we were just getting a value, we could also check if the query
	// was successfull but returned 0 rows with `if err == sql.ErrNoRows`.
	var createdAt, confirmedAt time.Time
	if u.ConfirmedAt > 0 {
		confirmedAt = time.Unix(u.ConfirmedAt, 0)
	} else {
		confirmedAt = time.Now()
	}
	err := db.GetInstance().QueryRow(qry, u.ID, u.Confirmed, confirmedAt).Scan(&u.ID, &u.Username, &u.Confirmed, &confirmedAt, &createdAt)

	if err != nil {
		logger.Errorf("%v:User:Update() Couldn't update user: %s", reqID, err.Error())
		dberr = db.FormError(err, qry, "users")
		return
	}

	u.CreatedAt = createdAt.Unix()
	if !confirmedAt.IsZero() {
		u.ConfirmedAt = confirmedAt.Unix()
	}
	u.Password = ""
	u.PasswordConfirm = ""
	u.Salt = ""
	return
}

// Delete deletes a user by ID
func Delete(db *database.DB, userID, reqID string) (dberr *database.Error) {
	logger.Debugf("User Delete ID: %s", userID)

	// id should always be first($1)!!!
	qry := `DELETE FROM users  WHERE id = $1`

	if _, err := db.GetInstance().Exec(qry, userID); err != nil {
		logger.Errorf("%v:User:Delete() Couldn't delete user: %s", reqID, err.Error())
		dberr = db.FormError(err, qry, "users")
		return
	}

	return
}

// scanAll scans a full row with all its columns into a user
func scanAll(row *sql.Row, u *User) error {
	var createdAt, confirmedAt time.Time
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Salt, &u.Confirmed, &confirmedAt, &createdAt)
	if err != nil {
		return err
	}
	u.CreatedAt = createdAt.Unix()
	if !confirmedAt.IsZero() {
		u.ConfirmedAt = confirmedAt.Unix()
	}
	return nil
}

// scanAllSafe scans a full row with all its columns into a user (except password related stuff)
func scanAllSafe(row *sql.Row, u *User) error {
	var createdAt, confirmedAt time.Time
	err := row.Scan(&u.ID, &u.Username, &u.Confirmed, &confirmedAt, &createdAt)
	if err != nil {
		return err
	}
	u.CreatedAt = createdAt.Unix()
	if !confirmedAt.IsZero() {
		u.ConfirmedAt = confirmedAt.Unix()
	}
	return nil
}

// formSetStatement generates a SET statement starting with $2
func formSetStatement(fields []string) string {
	// 	UPDATE films SET kind = 'Dramatic' WHERE kind = 'Drama';
	// SET temp_lo = temp_lo+1, temp_hi = temp_lo+15,
	if len(fields) == 0 {
		return ""
	}

	stmnt := "SET"
	i := 2
	for _, field := range fields {
		pair := fmt.Sprintf("%s = $%d", field, i)
		stmnt = fmt.Sprintf("%s %s,", stmnt, pair)
		i++
	}
	return stmnt[:len(stmnt)-1] // remove last ","
}
