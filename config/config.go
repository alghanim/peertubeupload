package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	APIConfig struct {
		URL      string `json:"url"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"apiConfig"`
	DBConfig struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Port     string `json:"port"`
		Host     string `json:"host"`
		Dbname   string `json:"dbname"`
	} `json:"dbConfig"`
}

func (c *Config) LoadConfiguration(file string) {
	configFile, err := os.Open(file)

	// If file not found, create a sample config file
	if os.IsNotExist(err) {
		*c = Config{
			APIConfig: struct {
				URL      string `json:"url"`
				Port     string `json:"port"`
				Username string `json:"username"`
				Password string `json:"password"`
			}{
				URL:      "localhost",
				Port:     "9000",
				Username: "root",
				Password: "ali12345",
			},
			DBConfig: struct {
				Username string `json:"username"`
				Password string `json:"password"`
				Port     string `json:"port"`
				Host     string `json:"host"`
				Dbname   string `json:"dbname"`
			}{
				Username: "postgres",
				Password: "password",
				Port:     "5432",
				Host:     "localhost",
				Dbname:   "postgres",
			},
		}
		configJSON, _ := json.MarshalIndent(*c, "", " ")
		_ = ioutil.WriteFile(file, configJSON, 0644)
	} else {
		defer configFile.Close()
		byteValue, _ := ioutil.ReadAll(configFile)
		json.Unmarshal(byteValue, c)
	}
}
