package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	APIConfig struct {
		URL      string `json:"url"`
		Port     string `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"apiConfig"`
	LoadType struct {
		LoadPathFromDB bool `json:"loadPathFromDB"`
		LoadFromFolder bool `json:"loadFromFolder"`
	} `json:"loadType"`
	FolderConfig struct {
		Path               string   `json:"path"`
		SpecificExtensions bool     `json:"specificextensions"`
		Extensions         []string `json:"extensions"`
	} `json:"folderConfig"`
	DBConfig struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Port     string `json:"port"`
		Host     string `json:"host"`
		Dbname   string `json:"dbname"`
	} `json:"dbConfig"`
	ProccessConfig struct {
		Threads int `json:"threads"`
	}
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
				URL:      "http://peertube.localhost",
				Port:     "9000",
				Username: "root",
				Password: "ali12345",
			},
			LoadType: struct {
				LoadPathFromDB bool `json:"loadPathFromDB"`
				LoadFromFolder bool `json:"loadFromFolder"`
			}{
				LoadPathFromDB: false,
				LoadFromFolder: true,
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
			FolderConfig: struct {
				Path               string   `json:"path"`
				SpecificExtensions bool     `json:"specificextensions"`
				Extensions         []string `json:"extensions"`
			}{
				Path:               "./videos/",
				SpecificExtensions: true,
				Extensions:         []string{".mp4", ".wmv"},
			},
			ProccessConfig: struct {
				Threads int `json:"threads"`
			}{
				Threads: 1,
			},
		}
		configJSON, _ := json.MarshalIndent(*c, "", " ")
		_ = os.WriteFile(file, configJSON, 0644)
		fmt.Println("No config file is there , created one for you .. please re-run the script after modifing the config.json file")
		os.Exit(0)
	} else {

		defer configFile.Close()
		byteValue, _ := io.ReadAll(configFile)
		err := json.Unmarshal(byteValue, c)
		if err != nil {
			panic("not able to read config.json ")
		}

	}
}
