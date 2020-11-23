package utils

import (
	"fmt"
	"time"

	"chocolate/service/api/shared/apierror"
	"chocolate/service/models/auth"
	"chocolate/service/shared/auth/jwt"
	"chocolate/service/shared/logger"
)

func GenerateAuthResponse(reqID, role, userID string, refreshExpiration time.Duration, eok bool) (authResponse *auth.Response, apierr *apierror.Error) {
	var (
		accessClaims, refreshClaims jwt.Claims
		accessToken, refreshToken   string
	)
	if accessClaims, apierr = GenerateAccessClaims(role, userID, eok); apierr != nil {
		return
	}
	logger.Debugf("%s:auth:GenerateToken() Access Claims: %+v:", reqID, accessClaims)
	if accessToken, apierr = jwt.Create(accessClaims); apierr != nil {
		return
	}
	logger.Debugf("%s:auth:GenerateToken() Access Token: %s:", reqID, accessToken)
	if refreshClaims, apierr = GenerateRefreshClaims(&accessClaims, refreshExpiration); apierr != nil {
		return
	}
	logger.Debugf("%s:auth:GenerateToken() Refresh Claims: %+v:", reqID, refreshClaims)
	if refreshToken, apierr = jwt.Create(refreshClaims); apierr != nil {
		return
	}
	logger.Debugf("%s:auth:GenerateToken() Refresh Token: %s:", reqID, refreshToken)

	authResponse = &auth.Response{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return
}

func GenerateAccessClaims(role, userID string, eok bool) (claims jwt.Claims, err *apierror.Error) {
	claims = jwt.New()
	now := time.Now()
	nowEpoch := now.Unix()
	exp := time.Hour * 24 * 7 //time.Second * 20 //
	claims.EmailOK = eok
	claims.ExpiresAt = now.Add(exp).Unix()
	claims.IssuedAt = nowEpoch
	claims.NotBefore = nowEpoch
	claims.UserID = userID
	claims.Role = role
	claims.Subject = fmt.Sprintf("/%ss/%s", role, userID)
	claims.TokenType = jwt.TokenTypeAccess
	return
}

func GenerateRefreshClaims(accessClaims *jwt.Claims, refreshExpiration time.Duration) (claims jwt.Claims, err *apierror.Error) {
	claims = jwt.New()
	now := time.Now()
	nowEpoch := now.Unix()

	claims.ExpiresAt = now.Add(refreshExpiration).Unix()
	claims.IssuedAt = nowEpoch
	claims.NotBefore = nowEpoch
	claims.UserID = accessClaims.UserID
	claims.Role = accessClaims.Role
	// The subject in this case is the Access_token ID
	claims.Subject = accessClaims.Id
	claims.TokenType = jwt.TokenTypeRefresh
	return
}
