package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"peertubeupload/config"
	"peertubeupload/database"
	"peertubeupload/model"
	"strings"
	"sync"

	"time"

	"golang.org/x/sync/semaphore"
)

var baseURL string
var client *http.Client
var c config.Config
var accessToken model.AccessToken

// var refreshToken string
var expirationTime time.Time
var tokenMutex = &sync.Mutex{}

func init() {
	c.LoadConfiguration("config.json")
	baseURL = fmt.Sprintf("%s:%s/api/v1", c.APIConfig.URL, c.APIConfig.Port)
}

func main() {

	log.SetFlags(log.Lshortfile)

	transport := &http.Transport{
		MaxConnsPerHost:     10,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	client = &http.Client{
		Timeout:   time.Minute * 10,
		Transport: transport,
	}

	loginClient, err := loginPrerequisite()
	if err != nil {
		panic(err)
	}
	// accessToken, err := login(loginClient, "password", c.APIConfig.Username, c.APIConfig.Password)
	// if err != nil {
	// 	panic(err)
	// }
	// // Set the new expiration time
	// expirationTime = time.Now().Add(time.Second * time.Duration(accessToken.ExpiresIn))

	filesChan := make(chan model.Media)
	ctx := context.Background()

	sem := semaphore.NewWeighted(int64(c.ProccessConfig.Threads))

	if c.LoadType.LoadFromFolder {

		go gatherPathsFromFolder(c.FolderConfig.Path, c.FolderConfig.Extensions, filesChan)

	} else if c.LoadType.LoadPathFromDB {
		db, err := database.InitDB(c)
		if err != nil {
			panic(err)
		}
		if db != nil {

			defer db.Close()
		}
		go gatherPathsFromDB(db, &c, c.FolderConfig.Extensions, filesChan)

	} else {
		fmt.Println("specify at least one load type .. ")
	}

	for f := range filesChan {
		if err := sem.Acquire(ctx, 1); err != nil {
			// handle the error
			log.Fatalf("Failed to acquire semaphore: %v", err)
		}
		go func(f model.Media) {
			defer sem.Release(1)
			// Process the file
			err = updateTokenIfNeeded(loginClient, "password", c.APIConfig.Username, c.APIConfig.Password)
			if err != nil {
				log.Println("Unable to get access token:", err)
				return
			}

			fileData, err := os.Stat(f.FilePath)
			if err != nil {
				log.Println(err)
			}

			video, err := uploadVideo(f.Title, "", fileData.ModTime().Format("2006-01-02 15:04:05"), accessToken.AccessToken, f.FilePath)
			if err != nil {
				fmt.Println()
			}
			jsonData, err := json.MarshalIndent(video, "", "    ")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(string(jsonData))
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
		log.Println(err)
		return
	}
	defer rows.Close()

	// Iterate through the rows and upload each video
	for rows.Next() {
		var media model.Media
		if err := rows.Scan(&media.Title, &media.Description, &media.FilePath); err != nil {
			fmt.Println(err)
			return
		}
		filechan <- media
	}
	close(filechan)
}
func gatherPathsFromFolder(root string, extensions []string, filesChan chan<- model.Media) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return nil
		}
		if !info.IsDir() {
			fileExt := strings.ToLower(filepath.Ext(info.Name()))
			for _, ext := range extensions {
				if ext == fileExt {
					filesChan <- model.Media{
						Title:       info.Name(),
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
		fmt.Printf("Error walking the path %q: %v\n", root, err)
	}
	close(filesChan)
}

func loginPrerequisite() (model.Login, error) {

	url := baseURL + "/oauth-clients/local"
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return model.Login{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return model.Login{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return model.Login{}, err
	}

	r, err := model.UnmarshalLogin(body)
	if err != nil {
		return model.Login{}, err
	}

	return r, nil

}

func login(loginClient model.Login, grant_type string, username string, password string) (model.AccessToken, error) {

	apiurl := baseURL + "/users/token"
	method := "POST"
	data := url.Values{
		"client_id":     {loginClient.ClientID},
		"client_secret": {loginClient.ClientSecret},
		"grant_type":    {grant_type},
		"response_type": {"code"},
		"username":      {username},
		"password":      {password},
	}

	req, err := http.NewRequest(method, apiurl, bytes.NewBufferString(data.Encode()))

	if err != nil {
		return model.AccessToken{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return model.AccessToken{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return model.AccessToken{}, err
	}

	accessToken, err := model.UnmarshalAccessToken(body)
	if err != nil {
		return model.AccessToken{}, err
	}
	if res.StatusCode != 200 {
		return model.AccessToken{}, fmt.Errorf("not authorized")
	}

	return accessToken, nil

}

func uploadVideo(title string, description string, originalDateTime string, token string, filePath string) (model.Video, error) {
	url := baseURL + "/videos/upload"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(filePath)
	if errFile1 != nil {
		return model.Video{}, errFile1
	}
	defer file.Close()
	part1, errFile1 := writer.CreateFormFile("videofile", filepath.Base(filePath))
	if errFile1 != nil {
		return model.Video{}, errFile1
	}
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {

		return model.Video{}, errFile1
	}
	_ = writer.WriteField("channelId", "1")
	_ = writer.WriteField("downloadEnabled", "false")
	_ = writer.WriteField("name", title)
	// _ = writer.WriteField("description", description)
	_ = writer.WriteField("commentsEnabled", "false")
	_ = writer.WriteField("originallyPublishedAt", originalDateTime)
	_ = writer.WriteField("privacy", "2")
	_ = writer.WriteField("waitTranscoding", "true")
	err := writer.Close()
	if err != nil {

		return model.Video{}, err
	}

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return model.Video{}, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return model.Video{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return model.Video{}, err
	}

	video, err := model.UnmarshalVideo(body)

	if err != nil {
		return model.Video{}, err
	}

	return video, nil
}

func updateTokenIfNeeded(loginClient model.Login, grant_type string, username string, password string) error {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	// Check if the current time is after the token expiration time
	if time.Now().After(expirationTime) {
		// Get a new token
		var err error
		if accessToken.RefreshToken == "" {
			// If we don't have a refresh token, do a full login
			accessToken, err = login(loginClient, grant_type, username, password)
		} else {
			// If we have a refresh token, use it to get a new access token
			accessToken, err = refreshAccessToken(loginClient, accessToken.RefreshToken)
		}
		if err != nil {
			return err
		}

		// Set the new expiration time
		expirationTime = time.Now().Add(time.Second * time.Duration(accessToken.ExpiresIn))
	}
	return nil
}
func refreshAccessToken(loginClient model.Login, refreshToken string) (model.AccessToken, error) {
	apiurl := baseURL + "/users/token"
	method := "POST"
	data := url.Values{
		"client_id":     {loginClient.ClientID},
		"client_secret": {loginClient.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequest(method, apiurl, bytes.NewBufferString(data.Encode()))

	if err != nil {
		return model.AccessToken{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return model.AccessToken{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return model.AccessToken{}, err
	}

	accessToken, err := model.UnmarshalAccessToken(body)
	if err != nil {
		return model.AccessToken{}, err
	}

	return accessToken, nil
}
