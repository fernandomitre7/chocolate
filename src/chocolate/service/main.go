package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"chocolate/service/database"
	"chocolate/service/shared/auth/jwt"
	"chocolate/service/shared/config"
	"chocolate/service/shared/email"
	"chocolate/service/shared/logger"
)

var (
	configFile = flag.String("config", "config/chocolate.conf.json", "Configuration File")
	logFile    = flag.String("logs", "logs/chocolate.log", "Log file Path")
	// Used for shutdown operations
	killC      chan os.Signal
	interruptC chan os.Signal
)

func createPIDFile(pid int) error {
	if pid > 0 {
		fmt.Println("Creating PID file")
		file, err := os.Create("bin/PID")
		if err != nil {
			return err
		}
		defer file.Close()
		file.WriteString(strconv.Itoa(pid))
		return nil
	}
	return errors.New("No PID to create file")
}

func deletePIDFile() error {
	fmt.Println("Deleting PID file")
	return os.Remove("bin/PID")
}

func getConfigFileLocation() (file string) {
	if file = os.Getenv("CHOCO_CONFIG"); file == "" {
		file = *configFile
	}
	return
}

func init() {
	// Initialize shutdown signals channels
	killC = make(chan os.Signal, 1)
	interruptC = make(chan os.Signal, 1)
}

func main() {
	flag.Parse()
	fmt.Println("############################################################################################")
	pid := os.Getpid()
	fmt.Println(fmt.Sprintf("API V1\tProcessID:%d\tDate:%v", pid, time.Now()))
	// Should I move it to init()??
	if err := createPIDFile(pid); err != nil {
		panic(err)
	}
	defer deletePIDFile()
	// Load Configurations
	configFileLoc := getConfigFileLocation()
	_conf, err := config.LoadConfiguration(configFileLoc)
	if err != nil {
		deletePIDFile()
		panic(err)
	}
	// Initialize Logger
	err = logger.Init(*logFile, _conf.Debug)
	if err != nil {
		deletePIDFile()
		panic(err)
	}

	logger.Debugf("API Starting on port %s:%s", _conf.Server.Host, _conf.Server.Port)

	time.Sleep(time.Second * 1)

	signal.Notify(killC, syscall.SIGTERM)
	signal.Notify(interruptC, os.Interrupt)

	var (
		server       *http.Server
		serverErrorC chan error
	)

	// Stop func to be called on system kill signals
	forceStop := func() {
		fmt.Println("\nStopping...")
		logger.Info("Stop, sending stop signal..")
		// Optionally, you could run server.Shutdown in a goroutine and block on
		// <-ctx.Done() if your application should wait for other services
		// to finalize based on context cancellation.
		close(serverErrorC)
		server.Shutdown(context.Background())
		return
	}
	// Stop function to be called when received error from server
	stop := func() {
		fmt.Println("\nStopping...")
		//server.Shutdown(context.Background())
		return
	}

	// Initialize DB
	var serviceDB *database.DB
	if serviceDB, err = database.New(_conf.DB); err != nil {
		panic(err)
	}
	defer serviceDB.Close()

	// Initialize Auth Keys
	if err = jwt.Init(_conf); err != nil {
		panic(err)
	}
	// Initialize Email Service
	email.Init(_conf.Email)

	// Start Server
	server, serverErrorC = startServer(_conf, serviceDB)
	logger.Debug("Http Server Started")

	// This select will prevent program to stop until is forcelly shutdown
	select {
	case <-interruptC:
		logger.Warn("INTERRUPT signal received! shutdown initiated...")
		forceStop()
	case <-killC:
		logger.Warn("SIGTERM signal received! shutdown initiated ...")
		forceStop()
	/* case <-processService.StoppedC:
	logger.Info("Received stopped signal from lower process")
	stop() */
	case serverError := <-serverErrorC:
		logger.Errorf("Got Server error %s, send signals to stop (if needed)", serverError.Error())
		stop()
	}

	logger.Info("API stopped")
	logger.Close()
	fmt.Println("Stopped.")
}

func startServer(conf *config.Configuration, serviceDB *database.DB) (*http.Server, chan error) {
	logger.Debug("Starting HTTP Server")

	errC := make(chan error)

	server := NewServer(conf, serviceDB)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Errorf("Error from http server %s", err.Error())
			errC <- err
		}
	}()

	return server, errC
}

// TODO: start https server
