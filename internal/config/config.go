package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"os"
)

// Config struct contains the server configuration
type Config struct {
	PermitPasswordLogin bool   `json:"PermitPasswordLogin"`
	PermitKeyLogin      bool   `json:"PermitKeyLogin"`
	PermitRootLogin     bool   `json:"PermitRootLogin"`
	AuthorizedKeysFile  string `json:"AuthorizedKeysFile"`
	PrivateKeyFile      string `json:"PrivateKeyFile"`
	UserKeysDir         string `json:"UserKeysDir"`
	EventsLogFile       string `json:"EventsLogFile"`
	SystemLogFile       string `json:"SystemLogFile"`
	AsyncEventsLog      bool   `json:"AsyncEventsLog"`
	AsyncSystemLog      bool   `json:"AsyncSystemLog"`
	ListenPort          int    `json:"ListenPort"`
	ListenAddress       string `json:"ListenAddress"`
}

//BastionConfig Hold the global configuration
var BastionConfig Config

// ParseConfig try to open and parse the file at the specified path.
// If the path is invalid or empty, the function will try to find a config file
// at the default locations.
func (c *Config) ParseConfig(path string) error {
	//TODO correct permissions
	if _, err := os.Stat("/var/lib/open-bastion/users/"); os.IsNotExist(err) {
		os.MkdirAll("/var/lib/open-bastion/users/", os.ModeDir)
	}

	if _, err := os.Stat("/var/lib/open-bastion/logs/"); os.IsNotExist(err) {
		os.MkdirAll("/var/lib/open-bastion/logs/", os.ModeDir)
	}

	home, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	// Default values if no configuration is provided for them
	defaultConfigPaths := []string{
		"/etc/open-bastion/open-bastion-conf.json",
		home + "/.config/open-bastion/open-bastion-conf.json",
		home + "/.config/open-bastion-conf.json",
		home + "/.open-bastion/open-bastion-conf.json",
	}

	defaultPrivateKey := home + "/.ssh/id_rsa"
	defaultAuthorizedKeys := home + "/.ssh/authorized_keys"
	defaultSSHPort := 22

	defaultLogsDirectory := "/var/lib/open-bastion/logs/"

	defaultUserKeysDirectory := "/var/lib/open-bastion/users/"

	defaultEventsLogFile := defaultLogsDirectory + "open-bastion-events.log"
	defaultSystemLogFile := defaultLogsDirectory + "open-bastion-system.log"

	configPath := ""

	if path != "" {
		_, err := os.Stat(path)

		if err != nil {
			//TODO add log message
		} else {
			configPath = path
		}
	}

	if configPath == "" {
		for _, p := range defaultConfigPaths {
			_, err := os.Stat(p)

			if err == nil {
				configPath = p
				break
			}
		}
	}

	if configPath == "" {
		return errors.New("Could not open any configuration file")
	}

	f, err := os.Open(configPath)

	if err != nil {
		return err
	}

	byteContent, err := ioutil.ReadAll(f)

	f.Close()

	if err != nil {
		return err
	}

	if !json.Valid(byteContent) {
		return errors.New("The configuration file is not a valid JSON file")
	}

	//By default, object keys which don't have a corresponding struct field are ignored
	err = json.Unmarshal([]byte(byteContent), &c)

	if err != nil {
		return err
	}

	if c.PermitPasswordLogin == false && c.PermitKeyLogin == false {
		return errors.New("No authorized login method")
	}

	if c.ListenPort == 0 {
		c.ListenPort = defaultSSHPort
	} else if c.ListenPort > 65535 || c.ListenPort < 0 {
		return errors.New("Invalid port configuration")
	}

	if net.ParseIP(c.ListenAddress) == nil {
		return errors.New("Invalid IP address configuration")
	}

	if c.PrivateKeyFile == "" {
		c.PrivateKeyFile = defaultPrivateKey
	} else {
		_, err := os.Stat(c.PrivateKeyFile)

		if err != nil {
			return errors.New("Private key file : invalid path")
		}
	}

	if c.AuthorizedKeysFile == "" {
		c.AuthorizedKeysFile = defaultAuthorizedKeys
	} else {
		_, err := os.Stat(c.AuthorizedKeysFile)

		if err != nil {
			return errors.New("Authorized keys file : invalid path")
		}
	}

	//TODO better log files verification
	if c.EventsLogFile == "" {
		c.EventsLogFile = defaultEventsLogFile
	}

	if c.SystemLogFile == "" {
		c.SystemLogFile = defaultSystemLogFile
	}

	if c.UserKeysDir == "" {
		c.UserKeysDir = defaultUserKeysDirectory
	}

	return nil
}
