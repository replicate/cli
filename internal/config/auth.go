package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL = "https://api.replicate.com/v1/"
)

var configFile string

type config map[string]Host

type Host struct {
	Token string `yaml:"token"`
}

func init() {
	// Look for config in the XDG_CONFIG_HOME directory
	if configDir, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists {
		configFile = filepath.Join(configDir, "replicate", "hosts")
	} else {
		// Look for config in the default directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			configFile = filepath.Join(homeDir, ".config", "replicate", "hosts")
		}
	}
}

func GetAPIBaseURL() string {
	url, found := os.LookupEnv("REPLICATE_BASE_URL")
	if found {
		return url
	}

	return DefaultBaseURL
}

func GetAPITokenForHost(host string) (string, error) {
	if host == "" {
		host = DefaultBaseURL
	}

	host, err := parseHost(host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %s", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	var c config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	if c == nil {
		return "", nil
	}

	h, ok := c[host]
	if !ok {
		return "", nil
	}

	return h.Token, nil
}

func GetAPIToken() (string, error) {
	return GetAPITokenForHost(GetAPIBaseURL())
}

func SetAPITokenForHost(apiToken, host string) error {
	if host == "" {
		host = DefaultBaseURL
	}

	host, err := parseHost(host)
	if err != nil {
		return fmt.Errorf("invalid host: %s", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(configFile), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		_, err = os.Create(configFile)
		if err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var c config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if c == nil {
		c = make(config)
	}

	c[host] = Host{Token: apiToken}

	data, err = yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config file: %w", err)
	}

	err = os.WriteFile(configFile, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func SetAPIToken(apiToken string) error {
	return SetAPITokenForHost(apiToken, GetAPIBaseURL())
}

func parseHost(host string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("Invalid host: %s", err)
	}
	return u.Hostname(), nil
}
