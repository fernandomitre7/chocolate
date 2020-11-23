package auth

import (
	"context"
	"net/http"
	"strings"

	"chocolate/service/api/shared/apierror"
	"chocolate/service/api/shared/responses"
	"chocolate/service/shared/auth/jwt"
	"chocolate/service/shared/logger"
	"chocolate/service/shared/reqcontext"
)

// Validate is the Validation Middleware that checks the request for Authorization Header
func Validate(next http.Handler, audience string, roles map[string]struct{}, checkEmail bool) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Debug("Validate:Checking Validations")

		var (
			authToken string
			err       *apierror.Error
			claims    *jwt.Claims
			ok        bool
		)
		// Get Token
		if authToken, err = extractAuthFromHeader(r); err != nil {
			responses.Error(r, rw, err)
			return
		}
		logger.Debugf("Auth Token: %s", authToken)
		// Verify JWT and get claims
		if claims, err = jwt.Verify(authToken); err != nil {
			responses.Error(r, rw, err)
			return
		}
		// Verify claims are correct
		if ok = claims.VerifyAudience(audience, true); !ok {
			err = apierror.New(http.StatusUnauthorized, "Wrong Audience in JWT", apierror.CodeUnauth)
			responses.Error(r, rw, err)
			return
		}
		if _, ok = roles[claims.Role]; !ok {
			err = apierror.New(http.StatusForbidden, "User not allowed to reach this endpoint", apierror.CodeForbidden)
			responses.Error(r, rw, err)
			return
		}

		// TODO: Move this claim verification to the respective /users route
		// We should add into the Route object a Middleware property, to allow
		// each route to define it respective middleware if necessary
		if claims.Role != jwt.RoleAdmin {
			if checkEmail && !claims.EmailOK {
				err = apierror.New(http.StatusForbidden, "Email not confirmed", apierror.CodeForbiddenNotConfirmed)
				// TODO: Add url for endpoint to resend confirmation email
				responses.Error(r, rw, err)
				return
			}
		}

		// TODO: Verify Token was not blacklisted (when user logsout)

		ctx := context.WithValue(r.Context(), reqcontext.AuthJWTKey, *claims)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func extractAuthFromHeader(r *http.Request) (string, *apierror.Error) {
	var err *apierror.Error

	authHeader := r.Header.Get("Authorization")
	logger.Debugf("Auth Header: %s", authHeader)

	if len(authHeader) == 0 {
		err = apierror.New(http.StatusUnauthorized, "Missing Authorization Header", "")
		return "", err
	}

	// Now checking whether a Bearer token can be found in an Authorization
	items := strings.Split(authHeader, " ")
	switch strings.ToLower(items[0]) {
	case "bearer":
		// Bearer should be followed by a token only.
		if len(items) != 2 {
			err = apierror.New(http.StatusUnauthorized, "Malformed Authorization Header", "")
			return "", err
		}

		return items[1], nil
	default:
		err = apierror.New(http.StatusUnauthorized, "Wrong type of Authorization Header", "")
		return "", err
	}
}
