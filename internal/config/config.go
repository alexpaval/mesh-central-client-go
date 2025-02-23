package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

var configPath = xdg.ConfigHome + "/mcc"
var configName = "meshcentral-client"
var configType = "json"

// constant containing default path to config file
var DefaultConfigPath = configPath + "/" + configName + "." + configType

func CreateConfig(server string, username string, password string) error {
	viper.Set("profiles", []map[string]interface{}{
		{
			"name":     "default",
			"server": 	server,
			"username": username,
			"password": password,
		},
	})

	viper.Set("default_profile", "default")

	// create directory if it does not exist
	path := filepath.Dir(viper.ConfigFileUsed())
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	return viper.WriteConfig()
}


// LoadConfig loads the configuration
func LoadConfig() error {
	viper.SetConfigType(configType)

	// if the config file has not been set, set it to default
	if viper.ConfigFileUsed() == "" {
		viper.SetConfigName(configName)
		viper.AddConfigPath(configPath)
	}

	viper.SetDefault("default_profile", "default")

	viper.SetDefault("profiles", []map[string]interface{}{
		{
			"name":     "default",
			"server": 	"",
			"username": "",
			"password": "",
		},
	})

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil

}

func GetConfigPath() string {
	return viper.ConfigFileUsed()
}

func GetConfigJSON() (*string, error) {
	configJSON, err := json.MarshalIndent(viper.AllSettings(), "", "  ")
	if err != nil {
		return nil, err
	}
	configString := string(configJSON)
	return &configString, nil
}

// SaveConfig saves the configuration
func SaveConfig() error {
	return viper.WriteConfig()
}
