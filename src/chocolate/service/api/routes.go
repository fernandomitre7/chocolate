package api

import (
	"bytes"
	"html/template"
	"net/http"

	"chocolate/service/api/handlers/auth"
	"chocolate/service/api/handlers/users"
	"chocolate/service/api/shared/apierror"
	"chocolate/service/api/shared/responses"
)

// Route describes an API Route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	Auth        *RouteAuth
	HandlerFunc http.HandlerFunc
}

// NewRoute creates a new route based on paramerters
func NewRoute(name, method, pattern string, auth *RouteAuth, handler http.HandlerFunc) Route {
	return Route{
		Name:        name,
		Method:      method,
		Pattern:     pattern,
		Auth:        auth,
		HandlerFunc: handler,
	}
}

// Routes is a collection of Route
type Routes []Route

// RouteAuth defines the authorization parameters of the route
type RouteAuth struct {
	// Roles this Route is authorized to serve, right now only a map[string]struct{}
	// In the future it should be a map[string]CRUD to see which crud operations this role is allowed
	Roles map[string]struct{}
	// Audience defines who in the API is supposed to get this type of request.
	// Not the User type, but the API type, (in case we have different api endpoints)
	//Audience string
	CheckEmail bool
}

// NewRouteAuth creates a new RouteAuth
func NewRouteAuth(roles []string, checkEmail bool) *RouteAuth {
	authRoles := make(map[string]struct{})
	for _, role := range roles {
		authRoles[role] = struct{}{}
	}
	return &RouteAuth{
		Roles:      authRoles,
		CheckEmail: checkEmail,
	}
}

// TestSomething is the handler to test any new functionality just do whatever there
func TestSomething(w http.ResponseWriter, r *http.Request) {
	// test email
	var apierr *apierror.Error

	t, err := template.ParseFiles("config/email-templates/confirmed.html")
	if err != nil {
		apierr = apierror.New(http.StatusInternalServerError, err.Error(), apierror.CodeInternal)
		responses.Error(r, w, apierr)
		return
	}

	buf := new(bytes.Buffer)
	data := struct{ Username string }{Username: "fernando@yogob.mx"}
	if err = t.Execute(buf, data); err != nil {
		apierr = apierror.New(http.StatusInternalServerError, err.Error(), apierror.CodeInternal)
		responses.Error(r, w, apierr)
		return
	}

	responses.HTML(r, w, buf)
}

var routes = Routes{
	// For Testing
	// NewRoute("Test", "POST", "/test", nil, TestSomething),
	// NewRoute("Test", "GET", "/test", nil, TestSomething),
	// Auth
	NewRoute(
		"Get Access Token",
		"POST", "/v1/tokens",
		nil, auth.GenerateTokens),
	NewRoute(
		"Get Refresh Token",
		"POST", "/v1/tokens/refresh",
		NewRouteAuth([]string{"user", "business", "admin"}, false),
		auth.RefreshTokens),
	// Users
	NewRoute(
		"Create User",
		"POST", "/v1/users",
		nil, users.Create),
	NewRoute(
		"Get Users",
		"GET", "/v1/users",
		NewRouteAuth([]string{"admin"}, false),
		users.Get),
	NewRoute(
		"Get User By ID",
		"GET", "/v1/users/{user_id}",
		NewRouteAuth([]string{"user", "admin"}, true),
		users.GetByID),
	NewRoute(
		"Update User By ID",
		"PUT", "/v1/users/{user_id}",
		NewRouteAuth([]string{"user", "admin"}, true),
		users.Update),
	NewRoute(
		"Delete User By ID",
		"DELETE", "/v1/users/{user_id}",
		NewRouteAuth([]string{"admin"}, false),
		users.Delete),
	// TODO: should confirm should just be a PUT /users/user_id?? maybe with a specific query_param??
	NewRoute(
		"Confirm User",
		"GET", "/v1/users/{user_id}/confirm",
		nil, users.Confirm),
}
