package users

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"chocolate/service/api/shared/apierror"
	"chocolate/service/api/shared/reqbody"
	"chocolate/service/api/shared/responses"
	"chocolate/service/database"
	"chocolate/service/models/auth"
	"chocolate/service/models/users"
	"chocolate/service/shared/auth/jwt"
	"chocolate/service/shared/auth/utils"
	"chocolate/service/shared/email"
	"chocolate/service/shared/email/templates/confirm"
	"chocolate/service/shared/logger"
	"chocolate/service/shared/reqcontext"
	"chocolate/service/shared/security"
)

// Create creates a new user in database
func Create(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	db := reqcontext.GetDB(r)

	logger.Debugf("%s:users:Create()", reqID)
	var (
		apierr *apierror.Error
		dberr  *database.Error
		user   = &users.User{}
	)
	if db == nil {
		logger.Errorf("%s:users:Create() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}
	// Get Body
	if apierr = reqbody.Read(r, user); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	// validate user
	if err := user.Valid(); err != nil {
		logger.Errorf("%s:users:Create() invalid user:%+v, err: %s", reqID, user, err.Error())
		apierr = apierror.New(http.StatusBadRequest, "invalid user", apierror.CodeBadRequestBody)
		responses.Error(r, w, apierr)
		return
	}

	// Check password confirmation
	if strings.Compare(user.Password, user.PasswordConfirm) != 0 {
		apierr = apierror.New(http.StatusBadRequest, "Password confirmation doens't match", apierror.CodeBadReqPasswordConfirm)
		responses.Error(r, w, apierr)
		return
	}

	// Hash Password
	if apierr = generatePassword(user); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	logger.Debugf("Got hash %s and salt %s", user.Password, user.Salt)

	if dberr = user.Insert(db, reqID); dberr != nil {
		logger.Errorf("%s:users:Create() Got error from Insert: err: %v", reqID, dberr)
		switch code := dberr.Code; code {
		case database.ErrorAlreadyExists:
			apierr = apierror.New(http.StatusBadRequest, fmt.Sprintf("User %s already exists: %s", user.Username, dberr.Error()), apierror.CodeBadRequestBody)
		case database.ErrorModelInvalid:
			apierr = apierror.New(http.StatusBadRequest, fmt.Sprintf("User not valid: %s", dberr.Error()), apierror.CodeBadRequestBody)
		case database.ErrorGeneric, database.ErrorExecute:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dberr.Error()), apierror.CodeInternalDB)
		}
		responses.Error(r, w, apierr)
		return
	}

	baseURL := reqcontext.GetBaseURL(r)
	if apierr = sendConfirmationEmail(user, baseURL, reqID); apierr != nil {
		// Should I delete recently created User?
		// Should I send the conf email on a separate goroutine??
		responses.Error(r, w, apierr)
		return
	}

	responses.Created(r, w, user, "/users")
	return

}

// Get gets all users
func Get(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	db := reqcontext.GetDB(r)
	var apierr *apierror.Error
	logger.Debugf("%s:users:Get()", reqID)

	if db == nil {
		logger.Errorf("%s:users:Get() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}

	usersList, dbErr := users.GetList(db, reqID)
	if dbErr != nil {
		logger.Errorf("%s:users:Get() Got error from Select: err: %v", reqID, dbErr)
		switch code := dbErr.Code; code {
		case database.ErrorModelInvalid:
			apierr = apierror.New(http.StatusBadRequest, fmt.Sprintf("User not valid: %s", dbErr.Error()), apierror.CodeBadRequestBody)
		case database.ErrorGeneric, database.ErrorExecute:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dbErr.Error()), apierror.CodeInternalDB)
		default:
			apierr = nil
		}
	}

	if apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	responses.Ok(r, w, usersList, "/users")
	return

}

// GetByID Returns a User by ID
func GetByID(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	vars := reqcontext.GetPathParams(r)
	logger.Debugf("%v:auth:GetByID() Starts vars= %v", reqID, vars)
	var (
		apierr *apierror.Error
		userID string
		claims jwt.Claims
	)
	// Get current accessClaims
	claims = reqcontext.GetAuthJWT(r)
	logger.Debugf("%v:auth:GetByID() claims %+v", reqID, claims)
	db := reqcontext.GetDB(r)
	if db == nil {
		logger.Errorf("%s:users:GetByID() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}
	// Get user_id
	if userID, apierr = getUserID(&claims, vars, reqID); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}
	logger.Debugf("%s:users:GetByID() Path User ID: %s, Claim UserID: %s", reqID, userID, claims.UserID)
	if claims.Role == jwt.RoleUser && userID != claims.UserID {
		apierr = apierror.New(http.StatusForbidden, "You can't modify this resource", apierror.CodeForbidden)
		responses.Error(r, w, apierr)
		return
	}
	user, dbErr := users.GetByID(db, userID, reqID)
	if dbErr != nil {
		logger.Errorf("%s:users:GetByID() Got error from Select: err: %v", reqID, dbErr)
		switch code := dbErr.Code; code {
		case database.ErrorGeneric, database.ErrorExecute:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dbErr.Error()), apierror.CodeInternalDB)
		default:
			apierr = nil
		}
	}

	if apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	responses.Ok(r, w, user, "/users/"+userID)

}

// Update Returns a User by ID
func Update(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	vars := reqcontext.GetPathParams(r)
	claims := reqcontext.GetAuthJWT(r)
	logger.Debugf("%v:auth:Update() Starts vars= %v", reqID, vars)
	var (
		apierr    *apierror.Error
		userID    string
		user      = &users.User{}
		returnJWT bool
	)
	db := reqcontext.GetDB(r)
	if db == nil {
		logger.Errorf("%s:users:Update() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}
	// Get user_id
	userID, apierr = getUserID(&claims, vars, reqID)
	logger.Debugf("%s:users:Update() Path User ID: %s, Claim UserID: %s", reqID, userID, claims.UserID)
	if claims.Role == jwt.RoleUser && userID != claims.UserID {
		apierr = apierror.New(http.StatusForbidden, "You can't modify this resource", apierror.CodeForbidden)
		responses.Error(r, w, apierr)
		return
	}
	// Get Body
	if apierr = reqbody.Read(r, user); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}
	user.ID = userID
	// Can't modify Confirmed state
	user.Confirmed = claims.EmailOK
	// validate user
	if err := user.Valid(); err != nil {
		logger.Errorf("%s:users:Update() invalid user:%+v, err: %s", reqID, user, err.Error())
		apierr = apierror.New(http.StatusBadRequest, "invalid user", apierror.CodeBadRequestBody)
		responses.Error(r, w, apierr)
		return
	}

	if dbErr := user.Update(db, reqID); dbErr != nil {
		logger.Errorf("%s:users:Update() Got error from Update: err: %v", reqID, dbErr)
		switch code := dbErr.Code; code {
		case database.ErrorModelInvalid:
			apierr = apierror.New(http.StatusBadRequest, fmt.Sprintf("User not valid: %s", dbErr.Error()), apierror.CodeBadRequestBody)
		case database.ErrorGeneric, database.ErrorExecute:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dbErr.Error()), apierror.CodeInternalDB)
		}

		responses.Error(r, w, apierr)
		return
	}

	if returnJWT { // huh?? is this for user confirmation??
		// Form Access and Refresh Tokens and return Response
		exp := time.Unix(claims.ExpiresAt, 0)
		iat := time.Unix(claims.IssuedAt, 0)
		refreshExp := exp.Sub(iat)
		var authResponse *auth.Response
		if authResponse, apierr = utils.GenerateAuthResponse(reqID, jwt.RoleUser, user.ID,
			refreshExp, user.Confirmed); apierr != nil {
			responses.Error(r, w, apierr)
			return
		}
		responses.Ok(r, w, authResponse, "/users/"+userID)
	} else {
		responses.Ok(r, w, user, "/users/"+userID)
	}

}

// Delete deletes a User from db
func Delete(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	vars := reqcontext.GetPathParams(r)
	logger.Debugf("%v:auth:Delete() Starts vars= %v", reqID, vars)
	var (
		apierr *apierror.Error
		userID string
	)
	db := reqcontext.GetDB(r)
	if db == nil {
		logger.Errorf("%s:users:Delete() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}
	// Get user_id
	userID, ok := vars["user_id"]
	if !ok {
		logger.Errorf("%s:users:Delete()  No User ID found in path", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "User ID in path cannot be retrieved", apierror.CodeInternal)
		responses.Error(r, w, apierr)
		return
	}

	if dbErr := users.Delete(db, userID, reqID); dbErr != nil {
		logger.Errorf("%s:users:Delete() Got error from Delete: err: %v", reqID, dbErr)
		switch code := dbErr.Code; code {
		case database.ErrorModelInvalid:
			apierr = apierror.New(http.StatusBadRequest, fmt.Sprintf("User not valid: %s", dbErr.Error()), apierror.CodeBadRequestBody)
		case database.ErrorGeneric, database.ErrorExecute:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dbErr.Error()), apierror.CodeInternalDB)
		}
	}

	if apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	responses.NoContent(r, w, "/users/"+userID)

}

// Confirm confirms email of specific user
func Confirm(w http.ResponseWriter, r *http.Request) {
	reqID := reqcontext.GetReqID(r)
	logger.Debugf("%v:auth:Confirm() Starts", reqID)

	var (
		apierr        *apierror.Error
		confirmToken  string
		confirmClaims *jwt.Claims
	)

	db := reqcontext.GetDB(r)
	if db == nil {
		logger.Errorf("%s:users:Confirm() Missing DB", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "Couldnt reach DB", apierror.CodeInternalDB)
		responses.Error(r, w, apierr)
		return
	}
	// Get user_id
	vars := reqcontext.GetPathParams(r)
	userID, ok := vars["user_id"]
	if !ok {
		logger.Errorf("%s:users:Confirm()  No User ID found in path", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "User ID in path cannot be retrieved", apierror.CodeInternal)
		responses.Error(r, w, apierr)
		return
	}
	logger.Debugf("%s:users:Confirm() User ID: %s", userID)

	// Get Confirm Token
	if confirmToken = r.FormValue("t"); len(confirmToken) == 0 {
		apierr = apierror.New(http.StatusBadRequest, "Missing 'token' query parameter", apierror.CodeBadRequestParams)
		responses.Error(r, w, apierr)
		return
	}

	// Verify Confirm JWT and get claims
	if confirmClaims, apierr = jwt.Verify(confirmToken); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}
	if userID != confirmClaims.UserID {
		if confirmClaims.TokenType != jwt.TokenTypeConfirm {
			apierr = apierror.New(http.StatusForbidden, "Token doesnt belong to user", apierror.CodeUnauth)
			responses.Error(r, w, apierr)
			return
		}
	}
	if confirmClaims.TokenType != jwt.TokenTypeConfirm {
		apierr = apierror.New(http.StatusForbidden, "This is not a confirm token", apierror.CodeUnauth)
		responses.Error(r, w, apierr)
		return
	}

	// Update User ID with email confirmed
	u := &users.User{ID: userID, Confirmed: true}
	if dberr := u.Update(db, reqID); dberr != nil {
		logger.Errorf("%s:users:Confirm() Got error from Update: err: %v", reqID, dberr)
		switch code := dberr.Code; code {
		case database.ErrorNoRows:
			apierr = apierror.New(http.StatusNotFound, "User is not registered", apierror.CodeResourceNotFound)
		/* case database.ErrorDB, database.ErrorExecute:
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dberr.Error()), apierror.CodeInternalDB) */
		default:
			apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", dberr.Error()), apierror.CodeInternalDB)
		}
	}
	if apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	// Return HTML response
	var page *bytes.Buffer
	if page, apierr = getConfirmedPage(u.Username); apierr != nil {
		responses.Error(r, w, apierr)
		return
	}

	responses.HTML(r, w, page)
	return
}

func getUserID(claims *jwt.Claims, vars map[string]string, reqID string) (userID string, apierr *apierror.Error) {
	var ok bool
	logger.Debugf("%s:users:getUserID() vars: %s", reqID, vars)
	if userID, ok = vars["user_id"]; !ok {
		logger.Errorf("%s:users:getUserID()  No User ID found in path", reqID)
		apierr = apierror.New(http.StatusInternalServerError, "User ID in path cannot be retrieved", apierror.CodeInternal)
		return
	}
	if userID == "this" {
		userID = claims.UserID
	} else if userID == "" {
		apierr = apierror.New(http.StatusBadRequest, "No User ID", apierror.CodeBadRequestParams)
	}
	return
}

func getConfirmedPage(username string) (buf *bytes.Buffer, apierr *apierror.Error) {
	// TODO: Move location to config file
	t, err := template.ParseFiles("data/pages/confirmed.html")
	if err != nil {
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Couldn't get confirmed page: %s", err.Error()), apierror.CodeInternal)
		return
	}
	buf = new(bytes.Buffer)
	data := struct{ Username string }{Username: username}
	if err = t.Execute(buf, data); err != nil {
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Couldn't parse confirmed page: %s", err.Error()), apierror.CodeInternal)
		return
	}
	return
}

func generatePassword(u *users.User) (apierr *apierror.Error) {
	password, err := security.GeneratePassword(u.Password)
	if err != nil {
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Couldn't generate password: %s", err.Error()), apierror.CodeInternal)
	}
	u.Password = password.Hash
	u.Salt = password.Salt
	return
}

// sendConfirmationEmail uses email provider to send the new account confirmation email
func sendConfirmationEmail(u *users.User, baseURL, reqID string) (apierr *apierror.Error) {
	var token, confURL string
	if token, apierr = generateConfirmationToken(u, reqID); apierr != nil {
		logger.Errorf("%s:Failed to generate confirmation token: %s", reqID, apierr.Error())
		return
	}
	if confURL, apierr = generateConfirmationURL(baseURL, u.ID, token); apierr != nil {
		logger.Errorf("%s:Failed to genearate confirmation url: %s", reqID, apierr.Error())
		return
	}

	sender, emailErr := email.NewSender()
	if emailErr != nil {
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Couldnt Create Email Sender: %s", emailErr.Error()), apierror.CodeInternalEmail)
		return
	}

	confTemplate := confirm.NewTemplate(u.Username, confURL)

	mail := &email.Email{
		Type:     email.HTMLEmail,
		Subject:  "Bienvenido a Zale",
		From:     "fernandomitre7@gmail.com",
		To:       u.Username,
		Template: confTemplate,
	}
	if emailErr = sender.Send(mail); emailErr != nil {
		apierr = apierror.New(http.StatusInternalServerError, fmt.Sprintf("Couldnt send email: %s", emailErr.Error()), apierror.CodeInternalEmail)
		return
	}

	return
}

func generateConfirmationURL(baseURL, userID, confToken string) (string, *apierror.Error) {
	confURL, err := url.Parse(baseURL)
	if err != nil {
		return "", apierror.New(http.StatusInternalServerError, "Couldn't send Confirmation email", apierror.CodeInternal)
	}

	logger.Debugf("generateConfirmationURL Encoded User ID %s", userID)
	confPath := path.Join(confURL.Path, "users", userID, "confirm")
	confURL.Path = confPath
	q := confURL.Query()
	q.Set("t", confToken)
	confURL.RawQuery = q.Encode()
	return confURL.String(), nil
}

func generateConfirmationToken(u *users.User, reqID string) (string, *apierror.Error) {
	claims := jwt.New()
	now := time.Now()
	nowEpoch := now.Unix()
	claims.EmailOK = false
	claims.ExpiresAt = 0
	claims.IssuedAt = nowEpoch
	claims.NotBefore = nowEpoch
	claims.UserID = u.ID
	claims.Role = jwt.RoleUser
	claims.Subject = fmt.Sprintf("/users/%s/confirm", u.ID)
	claims.TokenType = jwt.TokenTypeConfirm

	logger.Debugf("%s:users:generateConfirmationToken()  Claims: %+v:", reqID, claims)
	return jwt.Create(claims)
}
