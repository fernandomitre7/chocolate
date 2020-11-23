package security

import (
	"golang.org/x/crypto/bcrypt"
	"chocolate/service/shared/logger"
)

// Password object
type Password struct {
	Hash string
	Salt string
}

// GeneratePassword does exactly that
func GeneratePassword(password string) (*Password, error) {
	logger.Debug("Generating Password")
	// Generate SALT this will give us a 32 byte, base64 encoded output
	salt, err := GenerateRandomString(32)
	if err != nil {
		logger.Errorf("Couldn't generate password salt: %s", err.Error())
		return nil, err
	}
	saltedPassword := salt + password
	var hash string
	if hash, err = HashPassword(saltedPassword); err != nil {
		logger.Errorf("Couldn't generate password hash: %s", err.Error())
		return nil, err
	}
	return &Password{hash, salt}, nil
}

// HashPassword takes a string password and returns a hash of it
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks if passwords match
func CheckPasswordHash(password, salt, hash string) bool {
	logger.Debugf("CheckPasswordHash password: %s, salt: %s, hash: %s", password, salt, hash)
	saltedPassword := salt + password
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	logger.Debugf("CheckPasswordHash err: %+v", err)
	return err == nil
}
