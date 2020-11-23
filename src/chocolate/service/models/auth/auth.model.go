package auth

import (
	"encoding/json"
	"errors"
)

const (
	UserTypeClient   = "client"
	UserTypeAdmin    = "admin"
	UserTypeBusiness = "business"
)

type UserAuth struct {
	GrantType string `json:"grant_type"`
	UserType  string `json:"user_type"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Remember  bool   `json:"remember"`
}

// Valid validates that UserAuth fields are correct
func (a UserAuth) Valid() (err error) {
	if a.GrantType != "password" {
		err = errors.New("Wrong grant_type")
	}
	if a.UserType != UserTypeClient && a.UserType != UserTypeBusiness && a.UserType != UserTypeAdmin {
		err = errors.New("Invalid user_type")
	}
	return
}

// JSON returns the json bytes of the object
func (a UserAuth) JSON() ([]byte, error) {
	return json.Marshal(a)
}

// Decode Unmarshal bytes into UserAuth
func (a *UserAuth) Decode(data []byte) error {
	return json.Unmarshal(data, a)
}

// DecodeUserAuth converts json bytes into a UserAuth
func DecodeUserAuth(data []byte) (userAuth *UserAuth, err error) {
	userAuth = &UserAuth{}
	err = json.Unmarshal(data, userAuth)
	return
}

// AuthResponse object for authentication responses with jwt pairs
type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JSON returns the json bytes of the object
func (a Response) JSON() ([]byte, error) {
	return json.Marshal(a)
}

// AuthRefresh is the object for refresh token request
type Refresh struct {
	RefreshToken string `json:"refresh_token"`
}

// JSON returns the json bytes of the object
func (r Refresh) JSON() ([]byte, error) {
	return json.Marshal(r)
}

// Decode converts Unmarhsls bytes into a AuthRefresh
func (r *Refresh) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r Refresh) Valid() error {
	if len(r.RefreshToken) == 0 {
		return errors.New("Missing 'refresh_token'")
	}
	return nil
}

/* // DecodeAuthRefresh converts json bytes into a AuthRefresh
func DecodeAuthRefresh(data []byte) (authRefresh *Refresh, err error) {
	authRefresh = &Refresh{}
	err = json.Unmarshal(data, authRefresh)
	return
} */
