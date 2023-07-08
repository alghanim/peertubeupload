package main

import (
	"database/sql"
	"fmt"

	"os"
	"peertubeupload/auth"
	"peertubeupload/config"
	"peertubeupload/database"
	"peertubeupload/httpclient"
	"peertubeupload/logger"
	"peertubeupload/login"
	"peertubeupload/media"
	"peertubeupload/model"
)

var c config.Config
var baseURL string

func init() {
	c.LoadConfiguration("config.json")
	baseURL = fmt.Sprintf("%s:%s/api/v1", c.APIConfig.URL, c.APIConfig.Port)

}

func main() {

	var db *sql.DB
	client := httpclient.New()
	var loginManager auth.Authenticator = &login.LoginManager{}
	loginClient, err := loginManager.LoginPrerequisite(baseURL, client)
	if err != nil {
		logger.LogError(err.Error(), nil)
		os.Exit(1)
	}

	if c.LoadType.LoadFromFolder {

		filesChan := make(chan model.Media)

		media.ProcessFromFileSystem(c, filesChan, loginClient, client, loginManager)

	} else if c.LoadType.LoadPathFromDB {

		filesChan := make(chan map[string]interface{})
		db, err = database.InitDB(&c)
		if err != nil {
			panic(err)
		}

		if db != nil {
			defer db.Close()
		}

		media.ProcessFromDB(db, &c, filesChan, loginClient, client, loginManager)

	} else {
		logger.LogError("You need to specify at least one load type either db or file", nil)
		logger.LogError("App will exit, please check config.json under loadConfig section", nil)
		os.Exit(1)
	}

}
