package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"os"
	"path/filepath"
	"peertubeupload/auth"
	"peertubeupload/config"
	"peertubeupload/database"
	"peertubeupload/httpclient"
	"peertubeupload/logger"
	"peertubeupload/login"
	"peertubeupload/media"
	"peertubeupload/medialog"
	"peertubeupload/model"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

var baseURL string
var c config.Config

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

	filesChan := make(chan model.Media)
	ctx := context.Background()

	sem := semaphore.NewWeighted(int64(c.ProccessConfig.Threads))

	if c.LoadType.LoadFromFolder {

		if c.LoadType.LogType == "db" {
			db, err = database.InitDB(c)
			if err != nil {
				panic(err)
			}
			if db != nil {

				defer db.Close()
			}
		}
		go gatherPathsFromFolder(c.FolderConfig.Path, c.LoadType.Extensions, filesChan)

	} else if c.LoadType.LoadPathFromDB {
		db, err = database.InitDB(c)
		if err != nil {
			panic(err)
		}

		if db != nil {
			defer db.Close()
		}

		go gatherPathsFromDB(db, &c, c.LoadType.Extensions, filesChan)

	} else {
		logger.LogError("You need to specify at least one load type either db or file", nil)
		logger.LogError("App will exit, please check config.json under loadConfig section", nil)
		os.Exit(1)
	}

	for f := range filesChan {
		if err := sem.Acquire(ctx, 1); err != nil {
			// handle the error
			log.Fatalf("Failed to acquire semaphore: %v", err)
		}
		go func(f model.Media) {
			defer sem.Release(1)
			// Process the file

			err = loginManager.UpdateTokenIfNeeded(baseURL, client, loginClient, "password", c.APIConfig.Username, c.APIConfig.Password)
			if err != nil {
				logger.LogError("Unable to get access token", map[string]interface{}{"error": err})
				return
			}

			fileData, err := os.Stat(f.FilePath)
			if err != nil {
				logger.LogError("unable to get file data to retrive the original date, today date will be submitted", map[string]interface{}{"error": err})
				fileData = nil
			}

			var video model.Video

			if fileData != nil {
				video, err = media.UploadMedia(baseURL, client, f.Title, "", fileData.ModTime().Format("2006-01-02 15:04:05"), loginManager.GetAccessToken(), f.FilePath, &c)

			} else {
				video, err = media.UploadMedia(baseURL, client, f.Title, "", time.Now().Format("2006-01-02 15:04:05"), loginManager.GetAccessToken(), f.FilePath, &c)

			}
			if err != nil {
				logger.LogError("error uploading media", map[string]interface{}{"error": err, "file": f.FilePath})
				return
			}

			if c.LoadType.LogType == "db" {

				err = medialog.LogResultToDB(video, f, &c, db)
				if err != nil {
					logger.LogError("failed to log result in DB", map[string]interface{}{"error": err})
				}
			} else if c.LoadType.LogType == "file" {

				err := medialog.LogResultToFile(video, f, &c)
				if err != nil {
					logger.LogError("failed to log result in file", map[string]interface{}{"error": err})
				}

			} else if c.LoadType.LogType == "none" {
				logger.LogInfo("DONE UPLOADING ", map[string]interface{}{"file": f.FilePath})
			}

		}(f)
	}
	// Wait for all processing to complete
	if err := sem.Acquire(ctx, (int64(c.ProccessConfig.Threads))); err != nil {
		// handle the error
		log.Fatalf("Failed to acquire semaphore: %v", err)
	}

}

func gatherPathsFromDB(db *sql.DB, config *config.Config, extensions []string, filechan chan<- model.Media) {
	// Query the database for video details
	rows, err := db.Query(fmt.Sprintf("SELECT %s, %s, %s FROM %s",
		config.DBConfig.Title, config.DBConfig.Description, config.DBConfig.FilePath, config.DBConfig.TableName))
	if err != nil {
		logger.LogError("Failed to get paths from DB", map[string]interface{}{"error": err})
		os.Exit(1)
	}
	defer rows.Close()

	// Iterate through the rows and upload each video
	for rows.Next() {
		var media model.Media
		if err := rows.Scan(&media.Title, &media.Description, &media.FilePath); err != nil {
			logger.LogWarning("Not able to scan result", map[string]interface{}{"error": err})
			continue
		}
		filechan <- media
	}
	close(filechan)
}
func gatherPathsFromFolder(root string, extensions []string, filesChan chan<- model.Media) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.LogError("Error accessing path", map[string]interface{}{"Path": path, "error": err})
			return nil
		}
		if !info.IsDir() {
			fileExt := strings.ToLower(filepath.Ext(info.Name()))
			for _, ext := range extensions {
				if ext == fileExt {
					filesChan <- model.Media{
						Title:       media.GetFileName(path),
						Description: "",
						FilePath:    path,
					}
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		logger.LogError("Error walking the path", map[string]interface{}{"Path": root, "error": err})
	}
	close(filesChan)
}
