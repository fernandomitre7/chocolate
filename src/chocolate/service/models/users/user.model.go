package users

import (
	"encoding/json"
	"errors"
)

// User describes a User type User
type User struct {
	// Account related fields
	ID              string `json:"id,omitempty"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	PasswordConfirm string `json:"password_confirm,omitempty"`
	Salt            string `json:"-"`
	// Was Email confirmed?
	Confirmed   bool  `json:"confirmed,omitempty"`
	ConfirmedAt int64 `json:"confirmed_at,omitempty"`
	// Personal Info
	// Names          string `json:"names,omitempty"`
	// FirstLastName  string `json:"first_last_name,omitempty"`
	// SecondLastName string `json:"second_last_name,omitempty"`

	CreatedAt int64 `json:"created_at"`
}

// Users is a slice of User
type Users []User

/**
 * User Type Functions
 */

// JSON returns the json bytes of the object
func (u User) JSON() ([]byte, error) {
	return json.Marshal(u)
}

// Valid checks that the Citisen is safe for DB
func (u User) Valid() (err error) {
	if len(u.ID) == 0 && (len(u.Username) == 0 || len(u.Password) == 0) {
		return errors.New("Missing Username or Password")
	} else if len(u.ID) > 0 && (len(u.Username) > 0 || len(u.Password) > 0) {
		return errors.New("Can't Modify Username or Password")
	}
	return
}

// Decode takes data and Unmarshals it into itself
func (u *User) Decode(data []byte) (err error) {
	return json.Unmarshal(data, u)
}

// TODO: Check if we need it and not just Decode
// DecodeUser converts json bytes into a UserAuth
func DecodeUser(data []byte) (u *User, err error) {
	u = &User{}
	err = json.Unmarshal(data, u)
	return
}

/**
 * Users Type Functions
 */

// JSON returns the json bytes of the object
func (c Users) JSON() ([]byte, error) {
	return json.Marshal(c)
}

// Valid checks that the user is safe for DB
func (c Users) Valid() (err error) {
	return nil
}

// Decode takes data and Unmarshals it into itself
func (c *Users) Decode(data []byte) (err error) {
	return json.Unmarshal(data, c)
}
