package auth

import (
	"fmt"
	"net/http"
	"time"

	"chocolate/service/api/shared/apierror"
	"chocolate/service/api/shared/reqbody"
	"chocolate/service/api/shared/responses"
	"chocolate/service/database"
	"chocolate/service/models/auth"
	"chocolate/service/models/users"
	"chocolate/service/shared/auth/jwt"
	"chocolate/service/shared/auth/utils"
	"chocolate/service/shared/logger"
	"chocolate/service/shared/reqcontext"
	"chocolate/service/shared/security"
)

// GenerateTokens Creates a user AccessToken/RefreshToken pair
func GenerateTokens(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	logger.Debugf("%v:auth:GenerateToken() Starts", reqID)
	var (
		apierr       *apierror.Error
		userAuth     = &auth.UserAuth{}
		authResponse *auth.Response
	)
	db := reqcontext.GetDB(r)
	if db == nil {
		logger.Errorf("%s:auth:GenerateToken() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}

	// Get Body
	if apierr = reqbody.Read(r, userAuth); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	logger.Debugf("%s:auth:GenerateToken() Got UserAuth: %+v", reqID, userAuth)
	if err := userAuth.Valid(); err != nil {
		apierr = apierror.New(http.StatusBadRequest, err.Error(), apierror.CodeBadRequestBody)
		responses.Error(r, w, apierr)
		return
	}

	if userAuth.UserType == auth.UserTypeClient {
		if authResponse, apierr = formUserAuthResponse(db, userAuth.Remember,
			userAuth.Username, userAuth.Password, reqID); apierr != nil {
			logger.Errorf("%s:auth:GenerateToken() Error Forming Auth Response: %s", reqID, apierr.Error())
			responses.Error(r, w, apierr)
			return
		}
	} else {
		responses.NotImplemented(r, w, "/tokens")
		return
	}

	responses.Created(r, w, authResponse, "/tokens")
	return
}

func formUserAuthResponse(db *database.DB, remember bool, username, password, reqID string) (authResponse *auth.Response, apierr *apierror.Error) {
	// Get User by username in DB
	// TODO: users.GetByUsername
	user, dberr := users.GetBy(db, "username", username, reqID)
	if dberr != nil {
		logger.Errorf("%s:auth:GenerateTokens() Got error from Get User: err: %v", reqID, dberr)
		switch code := dberr.Code; code {
		case database.ErrorNoRows:
			apierr = apierror.New(http.StatusNotFound, fmt.Sprintf("Username %s is not registered", username), apierror.CodeResourceNotFound)
		/* case database.ErrorDB, database.ErrorExecute:
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dberr.Error()), apierror.CodeInternalDB) */
		default:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dberr.Error()), apierror.CodeInternalDB)
		}
	}
	if apierr != nil {
		return
	}
	logger.Debugf("%s: User = %+v", reqID, user)
	logger.Debugf("%s:Req Username = '%s', DB Username = '%s'", reqID, username, user.Username)
	// Validate Passwords
	if !security.CheckPasswordHash(password, user.Salt, user.Password) {
		apierr = apierror.New(http.StatusUnauthorized, "Wrong credentials", apierror.CodeUnauth)
		return
	}
	// Generate User Claims
	userType := auth.UserTypeClient
	userID := user.ID

	var refreshExpiration time.Duration
	if remember {
		refreshExpiration = time.Hour * 24 * 30
	} else {
		refreshExpiration = time.Hour * 24 * 8
	}

	return utils.GenerateAuthResponse(reqID, userType, userID, refreshExpiration, user.Confirmed)

}

// RefreshTokens generates a fresh pair of tokens based on original accesss_token
func RefreshTokens(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	logger.Debugf("%v:auth:RefreshTokens() Starts", reqID)
	var (
		apierr        *apierror.Error
		refreshReq    = &auth.Refresh{}
		authResponse  *auth.Response
		accessClaims  jwt.Claims
		refreshClaims *jwt.Claims
	)
	// Get Request Body
	if apierr = reqbody.Read(r, refreshReq); apierr != nil {
		responses.Error(r, w, apierr)
	}

	// Get current accessClaims
	accessClaims = reqcontext.GetAuthJWT(r)
	// Get refreshClaims
	if refreshClaims, apierr = jwt.Verify(refreshReq.RefreshToken); apierr != nil {
		logger.Errorf("%v:auth:RefreshTokens() Refresh Token not Valid", reqID)
		responses.Error(r, w, apierr)
		return
	}
	// Validate between tokens
	if accessClaims.TokenType != jwt.TokenTypeRefresh {
		apierr = apierror.New(http.StatusUnauthorized, "This is not a refresh token", apierror.CodeUnauth)
	} else if accessClaims.Id != refreshClaims.Subject {
		apierr = apierror.New(http.StatusUnauthorized, "Refesh Token doesn't match Access Token", apierror.CodeUnauth)
	} else if accessClaims.UserID != refreshClaims.UserID {
		apierr = apierror.New(http.StatusUnauthorized, "Refesh Token User ID doesn't match Access Token", apierror.CodeUnauth)
	}
	if apierr != nil {
		responses.Error(r, w, apierr)
		return
	}
	// TODO Check what happens if we request GetAuthJWT and there are no Claims in context
	role := accessClaims.Role
	userID := accessClaims.UserID
	eok := accessClaims.EmailOK
	exp := time.Unix(refreshClaims.ExpiresAt, 0)
	iat := time.Unix(refreshClaims.IssuedAt, 0)
	refreshExp := exp.Sub(iat)
	//refreshExp := fmt.Sprintf("%v", delta.Hours()/24)
	logger.Debugf("Refresh Expiration: %v", refreshExp.Hours())
	if authResponse, apierr = utils.GenerateAuthResponse(reqID, role, userID, refreshExp, eok); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}
	responses.Created(r, w, authResponse, "/tokens")
	return
}
