package media

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"peertubeupload/auth"
	"peertubeupload/config"
	"peertubeupload/logger"
	"peertubeupload/medialog"
	"peertubeupload/model"
	"strings"
	"time"

	"golang.org/x/sync/semaphore"
)

var baseURL string

func ProcessFromFileSystem(c config.Config, filesChan chan model.Media, loginClient *model.Login, client *http.Client, loginManager auth.Authenticator) {

	baseURL = fmt.Sprintf("%s:%s/api/v1", c.APIConfig.URL, c.APIConfig.Port)
	ctx := context.Background()

	sem := semaphore.NewWeighted(int64(c.ProccessConfig.Threads))
	go gatherPathsFromFolder(&c, filesChan)

	for f := range filesChan {
		if err := sem.Acquire(ctx, 1); err != nil {
			// handle the error
			log.Fatalf("Failed to acquire semaphore: %v", err)
		}
		go func(f model.Media) {
			defer sem.Release(1)
			// Process the file

			err := loginManager.UpdateTokenIfNeeded(baseURL, client, loginClient, "password", c.APIConfig.Username, c.APIConfig.Password)
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
				f.CreateDate = fileData.ModTime()
				// video, err = UploadMedia(baseURL, client, f.Title, "", fileData.ModTime().Format("2006-01-02 15:04:05"), loginManager.GetAccessToken(), f.FilePath, &c)
				video, err = UploadMediaInChunksOS(&c, f, loginManager.GetAccessToken())
			} else {
				f.CreateDate = time.Now()
				// video, err = UploadMedia(baseURL, client, f.Title, "", time.Now().Format("2006-01-02 15:04:05"), loginManager.GetAccessToken(), f.FilePath, &c)
				video, err = UploadMediaInChunksOS(&c, f, loginManager.GetAccessToken())
			}
			if err != nil {
				logger.LogError("error uploading media", map[string]interface{}{"error": err, "file": f.FilePath})
				return
			}

			if c.LoadType.LogType == "file" {

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

func ProcessFromDB(db *sql.DB, config *config.Config, filechan chan map[string]interface{}, loginClient *model.Login, client *http.Client, loginManager auth.Authenticator) {

	baseURL = fmt.Sprintf("%s:%s/api/v1", config.APIConfig.URL, config.APIConfig.Port)
	ctx := context.Background()
	var err error
	sem := semaphore.NewWeighted(int64(config.ProccessConfig.Threads))

	go gatherPathsFromDB(db, config, filechan)

	for f := range filechan {
		if err := sem.Acquire(ctx, 1); err != nil {
			// handle the error
			log.Fatalf("Failed to acquire semaphore: %v", err)
		}
		go func(f map[string]interface{}) {
			defer sem.Release(1)
			// Process the file

			err = loginManager.UpdateTokenIfNeeded(baseURL, client, loginClient, "password", config.APIConfig.Username, config.APIConfig.Password)
			if err != nil {
				logger.LogError("Unable to get access token", map[string]interface{}{"error": err})
				return
			}

			filePath := fmt.Sprintf("%v", f[config.DBConfig.FilePath])

			fileData, err := os.Stat(filePath)
			if err != nil {
				logger.LogError("unable to get file data to retrive the original date, today date will be submitted", map[string]interface{}{"error": err})
				fileData = nil
			}

			var video model.Video

			title := fmt.Sprintf("%v", f[config.DBConfig.Title])
			description := fmt.Sprintf("%v", f[config.DBConfig.Description])

			if fileData != nil {
				media := model.Media{
					Title:       title,
					Description: description,
					FilePath:    filePath,
					CreateDate:  fileData.ModTime(),
				}
				video, err = UploadMediaInChunksOS(config, media, loginManager.GetAccessToken())

				// video, err = UploadMedia(baseURL, client, title, "", fileData.ModTime().Format("2006-01-02 15:04:05"), loginManager.GetAccessToken(), filePath, config)

			} else {

				media := model.Media{
					Title:       title,
					Description: description,
					FilePath:    filePath,
					CreateDate:  time.Now(),
				}

				video, err = UploadMediaInChunksOS(config, media, loginManager.GetAccessToken())
				// video, err = UploadMedia(baseURL, client, title, "", time.Now().Format("2006-01-02 15:04:05"), loginManager.GetAccessToken(), filePath, config)

			}
			if err != nil {
				logger.LogError("error uploading media", map[string]interface{}{"error": err, "file": filePath})
				return
			}

			if config.LoadType.LogType == "db" {

				err = medialog.LogResultToDB(video, f, config, db, filePath)
				if err != nil {
					logger.LogError("failed to log result in DB", map[string]interface{}{"error": err})
				}

			} else if config.LoadType.LogType == "none" {
				logger.LogInfo("DONE UPLOADING ", map[string]interface{}{"file": filePath})
			}

		}(f)
	}
	// Wait for all processing to complete
	if err := sem.Acquire(ctx, (int64(config.ProccessConfig.Threads))); err != nil {
		// handle the error
		log.Fatalf("Failed to acquire semaphore: %v", err)
	}
}

func gatherPathsFromFolder(c *config.Config, filesChan chan<- model.Media) {

	err := filepath.Walk(c.FolderConfig.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.LogError("Error accessing path", map[string]interface{}{"Path": path, "error": err})
			return nil
		}
		if !info.IsDir() {
			if c.LoadType.SpecificExtensions {
				fileExt := strings.ToLower(filepath.Ext(info.Name()))
				for _, ext := range c.LoadType.Extensions {
					if ext == fileExt {
						title := strings.Replace(strings.TrimSuffix(GetFileName(path), filepath.Ext(GetFileName(path))), "_", " ", -1)
						
						filesChan <- model.Media{
							Title:       title,
							Description: "",
							FilePath:    path,
						}
						break
					}
				}
			} else {
				filesChan <- model.Media{
					Title:       GetFileName(path),
					Description: "",
					FilePath:    path,
				}
			}
		}
		return nil
	})
	if err != nil {
		logger.LogError("Error walking the path", map[string]interface{}{"Path": c.FolderConfig.Path, "error": err})
	}
	close(filesChan)
}

func gatherPathsFromDB(db *sql.DB, config *config.Config, filechan chan<- map[string]interface{}) {
	// Query the database for video details
	combinedColumns := append([]string{config.DBConfig.Title, config.DBConfig.Description, config.DBConfig.FilePath}, config.DBConfig.MediaIdentifier...)
	rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(combinedColumns, ","), config.DBConfig.TableName))
	// rows, err := db.Query(fmt.Sprintf("SELECT %s, %s, %s FROM %s",
	// 	config.DBConfig.Title, config.DBConfig.Description, config.DBConfig.FilePath, config.DBConfig.TableName))
	if err != nil {
		logger.LogError("Failed to get paths from DB", map[string]interface{}{"error": err})
		os.Exit(1)
	}
	defer rows.Close()
	values := make([]interface{}, len(combinedColumns))
	valuePtrs := make([]interface{}, len(combinedColumns))

	// Iterate through the rows and upload each video
	for rows.Next() {
		for i := 0; i < len(combinedColumns); i++ {
			valuePtrs[i] = &values[i]
		}
		err = rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatal(err)
		}
		row := make(map[string]interface{})
		for i, colName := range combinedColumns {
			val := valuePtrs[i].(*interface{})
			row[colName] = *val
		}
		// fmt.Println(row)
		// var mediaFile model.Media
		// if err := rows.Scan(&mediaFile.Title, &mediaFile.Description, &mediaFile.FilePath); err != nil {
		// 	logger.LogWarning("Not able to scan result", map[string]interface{}{"error": err})
		// 	continue
		// }

		if config.LoadType.SpecificExtensions {
			filePath := fmt.Sprintf("%v", row[config.DBConfig.FilePath])
			filename := filepath.Base(filePath)
			fileExt := strings.ToLower(filepath.Ext(filename))
			for _, ext := range config.LoadType.Extensions {
				if ext == fileExt {
					filechan <- row
					break
				}
			}
		} else {
			filechan <- row
		}

	}
	close(filechan)
}
