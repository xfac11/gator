package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DatabaseUrl     string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	return nil

}
func Read() (Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("Could not read file. Error: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Could not unmarshal data to Config struct. Error: %w", err)
	}

	return config, nil
}

func getConfigPath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Could not find home directory. Error: %w", err)
	}
	path := filepath.Join(homePath, ".gatorconfig.json")
	return path, nil
}
