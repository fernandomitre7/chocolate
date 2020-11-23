package reqcontext

// TODO: Instead of having it in utils, create a "shared" directory/package
import (
	"net/http"
	"time"

	"chocolate/service/database"
	"chocolate/service/shared/auth/jwt"
)

type contextKey int

const (
	// ReqIDKey is the context key to get the request id
	ReqIDKey contextKey = 0
	//BaseURL is the context key to retrieve the api base url
	BaseURLKey contextKey = 1
	// DbKey is the context key to retrieve db pointer
	DbKey contextKey = 2
	// StartTimeKey is the context key to get the request start time
	StartTimeKey contextKey = 3
	// AuthJWTKey is the context key to get the request Auth JWT
	AuthJWTKey contextKey = 4
	// EnvKey is the context key to get the current server environment
	EnvKey contextKey = 5
	// PathParams key to get the gorilla mux vars = path params
	PathParamsKey contextKey = 6
)

// GetReqID return the Request ID
func GetReqID(r *http.Request) string {
	return r.Context().Value(ReqIDKey).(string)
}

// GetStartTime returns the StartTime of the http request
func GetStartTime(r *http.Request) time.Time {
	return r.Context().Value(StartTimeKey).(time.Time)
}

// GetAuthJWT gets the current request user claims
func GetAuthJWT(r *http.Request) jwt.Claims {
	return r.Context().Value(AuthJWTKey).(jwt.Claims)
}

// GetEnvironment gets server business environment
func GetEnvironment(r *http.Request) string {
	return r.Context().Value(EnvKey).(string)
}

// IsProduction directly checks environment to see if it is production
func IsProduction(r *http.Request) bool {
	return GetEnvironment(r) == "production"
}

// GetDB returns the db pointer
func GetDB(r *http.Request) *database.DB {
	return r.Context().Value(DbKey).(*database.DB)
}

// GetBaseURL returns the API Server Base URL
func GetBaseURL(r *http.Request) string {
	return r.Context().Value(BaseURLKey).(string)
}

// GetPathParams gets the route path params
func GetPathParams(r *http.Request) map[string]string {
	return r.Context().Value(PathParamsKey).(map[string]string)
}
