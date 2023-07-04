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
		LoadPathFromDB bool     `json:"loadPathFromDB"`
		LoadFromFolder bool     `json:"loadFromFolder"`
		Extensions     []string `json:"extensions"`
	} `json:"loadType"`
	FolderConfig struct {
		Path               string `json:"path"`
		SpecificExtensions bool   `json:"specificextensions"`
	} `json:"folderConfig"`
	DBConfig struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Port        string `json:"port"`
		Host        string `json:"host"`
		Dbname      string `json:"dbname"`
		TableName   string `json:"table_name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		FilePath    string `json:"file_path"`
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
				LoadPathFromDB bool     `json:"loadPathFromDB"`
				LoadFromFolder bool     `json:"loadFromFolder"`
				Extensions     []string `json:"extensions"`
			}{
				LoadPathFromDB: false,
				LoadFromFolder: true,
				Extensions:     []string{".mp4", ".wmv"},
			},
			DBConfig: struct {
				Username    string `json:"username"`
				Password    string `json:"password"`
				Port        string `json:"port"`
				Host        string `json:"host"`
				Dbname      string `json:"dbname"`
				TableName   string `json:"table_name"`
				Title       string `json:"title"`
				Description string `json:"description"`
				FilePath    string `json:"file_path"`
			}{
				Username:    "postgres",
				Password:    "password",
				Port:        "5432",
				Host:        "localhost",
				Dbname:      "postgres",
				TableName:   "media_table",
				Title:       "title_column",
				Description: "description_column",
				FilePath:    "file_path_column",
			},
			FolderConfig: struct {
				Path               string `json:"path"`
				SpecificExtensions bool   `json:"specificextensions"`
			}{
				Path:               "./videos/",
				SpecificExtensions: true,
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
