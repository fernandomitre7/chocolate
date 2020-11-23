package jwt

import (
	"errors"

	"chocolate/service/shared/logger"
	"chocolate/service/shared/utils/uuid"

	_jwt "github.com/dgrijalva/jwt-go"
)

const (
	// Token Types
	TokenTypeAccess  = "access_token"
	TokenTypeRefresh = "refresh_token"
	TokenTypeConfirm = "confirm_token"
	// AuthType
	AuthTypeBearer = "bearer"
	// Roles
	RoleUser     = "user"
	RoleBusiness = "business"
	RoleAdmin    = "admin"
)

// Claims extends the StandarClaims
type Claims struct {
	_jwt.StandardClaims
	// UserID is the User ID
	UserID string `json:"uid"`
	// EmailOK Email was confirmed i.e. user is valid
	EmailOK bool `json:"eok"`
	// Role user role "admin"|"user"|"business"
	Role string `json:"rol"`
	// TokenType is either an access_token, refresh_token, confirm_token
	TokenType string `json:"ttp"`
	// AuthType is the type of auth for the JWT (for now always "bearer")
	AuthType string `json:"ath"`
}

// New Creates New set of JWT Claims
func New() Claims {
	c := Claims{}
	c.Audience = audience
	c.Issuer = audience
	c.Id, _ = uuid.New()
	c.AuthType = AuthTypeBearer
	return c
}

// Valid is called by JWT Parser method
func (c Claims) Valid() error {

	// This checks for expiration, issuedat and notbefore
	if ve := c.StandardClaims.Valid(); ve != nil {
		logger.Infof("Invalid Standard Claims: %v", ve.Error())
		return ve
	}

	var inner error
	if c.Role == "" {
		inner = errors.New("Missing rol claim")
	} else if c.UserID == "" {
		inner = errors.New("Missing uid claim")
	}

	if inner != nil {
		ve := new(_jwt.ValidationError)
		ve.Errors = _jwt.ValidationErrorMalformed
		return ve
	}

	return nil
}
