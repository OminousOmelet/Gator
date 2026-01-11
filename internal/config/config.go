package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const CONFIG_FILE_NAME = ".gatorconfig.json"

func Read() (*Config, error) {
	var cfg Config
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return &cfg, err
	}

	jsonData, err := os.ReadFile(homeDir + "/" + CONFIG_FILE_NAME)
	if err != nil {
		return &cfg, err
	}

	if err := json.Unmarshal(jsonData, &cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user

	jsonData, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}

	os.WriteFile(homeDir+"/"+CONFIG_FILE_NAME, jsonData, 0777)
	return nil
}
