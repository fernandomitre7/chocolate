package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Configuration is the structure holding projects configuration environment variables
type Configuration struct {
	Server      ServerConfig     `json:"server"`
	Environment string           `json:"env"`
	Debug       bool             `json:"debug"`
	JWT         jwtConfiguration `json:"jwt"`
	DB          SQLConfig        `json:"db"`
	Email       EmailConfig      `json:"email"`
}

// ServerConfig holds all the server configurations
type ServerConfig struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Protocol   string `json:"protocol"`
	APIVersion string `json:"api_version"`
}

type jwtConfiguration struct {
	PubKey   string `json:"pub_key"`
	PrivKey  string `json:"priv_key"`
	Audience string `json:"audience"`
}

// SQLConfig holds the configuration used for instantiating a new SQL DB.
type SQLConfig struct {
	// Address that locates our postgres instance
	Host string `json:"host"`
	// Port to connect to
	Port string `json:"port"`
	// User that has access to the database
	User string `json:"user"`
	// Password so that the user can login
	Password string `json:"password"`
	// Database to connect to (must have been created priorly)
	Database string `json:"database"`
}

// EmailConfig hodls the configuration used for email sending
type EmailConfig struct {
	Provider  string            `json:"provider"`
	Host      string            `json:"host"`
	Port      string            `json:"port"`
	TLS       bool              `json:"tls"`
	Auth      emailAuthConfig   `json:"auth"`
	Templates map[string]string `json:"templates"`
}
type emailAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	_conf *Configuration
)

// LoadConfiguration loads project configuration from a file and from specified env vars
func LoadConfiguration(filePath string) (conf *Configuration, err error) {
	fmt.Printf("Loading Config file: %v \n", filePath)
	var file *os.File
	if file, err = os.Open(filePath); err != nil {
		fmt.Printf("[EROR] Error openning config file %s: %v \n", filePath, err.Error())
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&conf); err != nil {
		fmt.Printf("[EROR] Error decoding config file %v \n", err.Error())
		return nil, err
	}
	_conf = conf
	return

	// TODO: Set into Configuration the values that are stored in env vars
	//	 Values in env vars are for security, passwords and stuff
	//_conf.DB.Database = os.Getenv("Connection_String")
}

// GetConfiguration returns the loaded configuration
func GetConfiguration() (conf *Configuration) {
	return _conf
}
