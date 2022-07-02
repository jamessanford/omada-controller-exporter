package omada

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Path     string `yaml:"path"`
	Username string `yaml:"user"`
	Password string `yaml:"pass"`
	Secure   bool   `yaml:"secure"` // control tls.Config InsecureSkipVerify
}

var errMissingConfig = errors.New("ensure OMADA_PATH OMADA_USER OMADA_PASS are configured or use a config file")

// ParseConfig reads a config file, if available, and then
// uses values from these environment variables:
//
//     OMADA_PATH - base URL to Omada controller, like "https://192.168.255.123:8043/"
//     OMADA_USER - controller admin username
//     OMADA_PASS - controller admin password
//     OMADA_SECURE - require verifiable TLS connection to OMADA_PATH when 1 or true
//
// Note that "secure" is false by default, the default controller does not
// have a verifiable TLS certificate.
func ParseConfig(configFile string) (*Config, error) {
	config := &Config{}

	if configFile != "" {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("config %q: %w", configFile, err)
		}
		if err = yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("config %q: %w", configFile, err)
		}
	}

	if path, ok := os.LookupEnv("OMADA_PATH"); ok {
		config.Path = path
	}
	if user, ok := os.LookupEnv("OMADA_USER"); ok {
		config.Username = user
	}
	if pass, ok := os.LookupEnv("OMADA_PASS"); ok {
		config.Password = pass
	}
	if secure, ok := os.LookupEnv("OMADA_SECURE"); ok {
		switch {
		case secure == "0" || strings.ToLower(secure) == "false":
			config.Secure = false
		case secure == "1" || strings.ToLower(secure) == "true":
			config.Secure = true
		default:
			return nil, fmt.Errorf("OMADA_SECURE has invalid value %q", secure)
		}
	}

	if config.Path == "" || config.Username == "" || config.Password == "" {
		return nil, errMissingConfig
	}

	return config, nil
}
