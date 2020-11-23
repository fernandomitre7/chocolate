package jwt

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"

	"chocolate/service/api/shared/apierror"
	"chocolate/service/shared/config"
	"chocolate/service/shared/logger"

	_jwt "github.com/dgrijalva/jwt-go"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	audience  string
)

// Init initializes jwt, loads public and private key
func Init(conf *config.Configuration) (err error) {
	var signBytes, verifyBytes []byte

	logger.Debugf("jwt:Init() Audience: %s, PrivKey: %s, PubKey: %s", conf.JWT.Audience, conf.JWT.PrivKey, conf.JWT.PubKey)

	audience = conf.JWT.Audience

	if signBytes, err = ioutil.ReadFile(conf.JWT.PrivKey); err != nil {
		logger.Errorf("Error loading JWT Private Key File: %s", err.Error())
		return
	}
	if verifyBytes, err = ioutil.ReadFile(conf.JWT.PubKey); err != nil {
		logger.Errorf("Error loading JWT Public Key File: %s", err.Error())
		return
	}
	//openssl genrsa -f4 -out jwt_key.priv 4096
	if signKey, err = _jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
		logger.Errorf("Error parsing JWT private Key: %s", err.Error())
		return
	}
	//openssl rsa -in jwt_key.priv -outform PEM -pubout -out jwt_key.pub
	if verifyKey, err = _jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
		logger.Errorf("Error parsing JWT public Key: %s", err.Error())
		return
	}
	return
}

// Verify parses token to see if is a valid JWT
// It only validates "standard" JWT claims, we still need to validate Zale claims
func Verify(jwtStr string) (*Claims, *apierror.Error) {
	logger.Debugf("jwt:Verify() JWT: %s", jwtStr)
	var (
		ve     *_jwt.ValidationError
		apierr *apierror.Error
	)

	// We parse JWT using Zale's user Claims
	token, err := _jwt.ParseWithClaims(jwtStr, &Claims{}, func(token *_jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*_jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return verifyKey, nil
	})

	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		logger.Debug("Valid JWT")
		return claims, nil
	}

	apierr = apierror.New(http.StatusUnauthorized, "Unauthorized", apierror.CodeUnauth)
	// Check what type of error was
	if ve, ok = err.(*_jwt.ValidationError); ok {
		if ve.Errors&_jwt.ValidationErrorMalformed != 0 {
			logger.Infof("Malformed Token: %v", ve)
			apierr = apierror.New(http.StatusUnauthorized, "Malformed JWT", apierror.CodeUnauthMalformed)
		} else if ve.Errors&_jwt.ValidationErrorExpired != 0 {
			logger.Infof("Expired Token: %v", ve)
			apierr = apierror.New(http.StatusUnauthorized, "Expired JWT", apierror.CodeUnauthExpired)
		} else if ve.Errors&_jwt.ValidationErrorNotValidYet != 0 {
			logger.Infof("Token Not Valid Yet: %v", ve)
			apierr = apierror.New(http.StatusUnauthorized, "JWT Not Valid Yet", apierror.CodeUnauthNotActive)
		}
	}

	return nil, apierr

}

// Create generates a new JWT, returns it as a string
func Create(claims Claims) (string, *apierror.Error) {
	token := _jwt.NewWithClaims(_jwt.SigningMethodRS256, claims)
	jwtStr, err := token.SignedString(signKey)
	if err != nil {
		logger.Errorf("Error creating JWT: %v", err)
		return "", apierror.New(http.StatusInternalServerError, "Coulnd't Generate JWT: "+err.Error(), apierror.CodeInternalJWT)
	}
	return jwtStr, nil
}
