package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"chocolate/service/api/metrics"
	"chocolate/service/database"
	"chocolate/service/shared/auth"
	"chocolate/service/shared/config"
	"chocolate/service/shared/logger"
	"chocolate/service/shared/reqcontext"
	"chocolate/service/shared/utils/uuid"
)

// NewRouter creates a new Router, gets the routes and registers them in router
func NewRouter(conf *config.Configuration, apidb *database.DB) *mux.Router {

	audience := conf.JWT.Audience

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		handler := route.HandlerFunc

		// Middleware is executed in the reverse on how it is added

		handler = metrics.Log(handler, route.Name)

		if route.Auth != nil {
			// Assign Authorization Validation
			handler = auth.Validate(handler, audience, route.Auth.Roles, route.Auth.CheckEmail)
		}

		handler = addContext(handler, conf, apidb)

		logger.Debugf("Installing route '%s' for pattern '%s' with method '%s'", route.Name, route.Pattern, route.Method)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

func addContext(next http.Handler, conf *config.Configuration, apidb *database.DB) http.HandlerFunc {
	// Get necessary env vars we might need to pass to the req
	env := conf.Environment
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		pathParams := mux.Vars(r)
		reqID, err := uuid.New()
		if err != nil {
			logger.Errorf("Error While creating Request ID UUID: %s", err.Error())
		}
		logger.Debugf("%s:Path Params: %v", reqID, pathParams)
		startTime := time.Now()
		ctx := context.WithValue(context.Background(), reqcontext.StartTimeKey, startTime)
		ctx = context.WithValue(ctx, reqcontext.PathParamsKey, pathParams)
		ctx = context.WithValue(ctx, reqcontext.BaseURLKey, apiRoutesBaseURL(conf.Server))
		ctx = context.WithValue(ctx, reqcontext.DbKey, apidb)
		ctx = context.WithValue(ctx, reqcontext.ReqIDKey, reqID)
		ctx = context.WithValue(ctx, reqcontext.EnvKey, env)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func apiRoutesBaseURL(serverConf config.ServerConfig) string {
	baseURL := &url.URL{}
	baseURL.Scheme = serverConf.Protocol
	baseURL.Host = fmt.Sprintf("%s:%s", serverConf.Host, serverConf.Port)
	baseURL.Path = serverConf.APIVersion
	logger.Debugf("API Base URL %s", baseURL.String())
	return baseURL.String()
}
