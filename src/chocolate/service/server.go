package main

import (
	"fmt"
	"net/http"
	"time"

	"chocolate/service/database"

	"chocolate/service/api"
	"chocolate/service/shared/config"
)

// NewServer creates a new serverd depending on configurations
func NewServer(conf *config.Configuration, serviceDB *database.DB) *http.Server {
	// Get Router
	r := api.NewRouter(conf, serviceDB)

	addr := fmt.Sprintf("%s:%s", conf.Server.Host, conf.Server.Port)

	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Minute * 2,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Minute * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	return srv
}

// TODO: implement an HTTPS Server
func NewSecureServer() {
}
