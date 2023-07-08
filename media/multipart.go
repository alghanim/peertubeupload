package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"peertubeupload/config"
	"peertubeupload/logger"
	"peertubeupload/model"
	"strings"
	"time"
)

/*
This file contains code originally written by FiskFan1999, licensed under the MIT License.
The original code can be found at: https://github.com/FiskFan1999/peertube-multipart-upload
*/

type PeertubeUserToken struct {
	Access_token string
}
type OauthClientsLocal struct {
	Client_id     string
	Client_secret string
}
type MultipartUploadHandlerHandlerInput struct {
	Hostname              string
	Username              string
	Password              string
	ContentType           string
	ChannelID             int
	File                  *VideoFileReader
	FileName              string
	DisplayName           string
	Privacy               int8
	Category              int
	CommentsEnabled       bool
	DescriptionText       string
	DownloadEnabled       bool
	Language              string
	Licence               int
	NSFW                  bool
	SupportText           string
	Tags                  []string
	OriginallyPublishedAt string
}

func MultipartUploadHandler(input MultipartUploadHandlerHandlerInput, token string) (video model.Video, err error) {

	client := &http.Client{}
	initializeUrl := fmt.Sprintf("%s/api/v1/videos/upload-resumable", input.Hostname)
	initializePayload := map[string]interface{}{
		"channelId":             input.ChannelID,
		"filename":              input.FileName,
		"name":                  input.DisplayName,
		"commentsEnabled":       input.CommentsEnabled,
		"downloadEnabled":       input.DownloadEnabled,
		"privacy":               input.Privacy,
		"waitTranscoding":       true,
		"originallyPublishedAt": input.OriginallyPublishedAt,
	}
	initializePayloadBytes, err := json.Marshal(initializePayload)
	if err != nil {
		panic(err)
	}
	initialize, err := http.NewRequest("POST", initializeUrl, bytes.NewReader(initializePayloadBytes))
	if err != nil {
		panic(err)
	}

	initialize.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	initialize.Header.Add("X-Upload-Content-Length", fmt.Sprintf("%d", input.File.TotalBytes))
	initialize.Header.Add("X-Upload-Content-Type", input.ContentType)
	initialize.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(initialize)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 201 {
		logger.LogInfo("initialize api call returned status code ", map[string]interface{}{"status code": resp.StatusCode})

		_, err2 := io.ReadAll(resp.Body)
		if err2 != nil {
			panic(err)
		}
		defer resp.Body.Close()

		panic("returned non 201 status code")
	}

	defer resp.Body.Close()
	uploadLocation := resp.Header.Get("Location")

	if strings.HasPrefix(uploadLocation, "//") {
		uploadLocation = "http:" + uploadLocation
	} else {
		logger.LogWarning("Warning: recieved an upload location that doesn't begin with \"//\", i don't know what to do with this.", map[string]interface{}{"file": input.FileName})
		return
	}
	logger.LogInfo("Updload Location", map[string]interface{}{"location": uploadLocation})

	for {
		chunk, err := input.File.GetNextChunk()
		if err != nil {
			logger.LogError("error getting next chunk", map[string]interface{}{"error": err, "file": input.FileName})
			break
		}
		if chunk.Finished {
			break
		}

		// logger.LogInfo("upload details", map[string]interface{}{"MinBye": chunk.MinByte, "MaxByte": chunk.MaxByte, "length": chunk.Length, "RangeHeader": chunk.RangeHeader})

		for {
			up, err := http.NewRequest("PUT", uploadLocation, bytes.NewReader(chunk.Bytes))
			if err != nil {
				panic(err)
			}

			up.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
			up.Header.Add("Content-Length", fmt.Sprintf("%d", chunk.Length))
			up.Header.Add("Content-Range", chunk.RangeHeader)

			resp, err := client.Do(up)
			if err != nil {
				panic(err)
			}
			// logger.LogInfo(resp.Status, nil)
			defer resp.Body.Close()
			body, err2 := io.ReadAll(resp.Body)
			if err2 != nil {
				panic(err2)
			}
			if len(body) != 0 {
				video, err = model.UnmarshalVideo(body)

				if err != nil {
					logger.LogError("Unable to marshal result", map[string]interface{}{"error": err})
					return video, err
				}

			}

			// fmt.Println(string(body))
			if resp.StatusCode == 308 || resp.StatusCode == 200 {
				break
			} else {
				logger.LogWarning("Status code other than 308 or 200 recieved. Will retry.", nil)
				time.Sleep(15 * time.Second)
			}
		}
	}
	return video, nil
}
func UploadMediaInChunksOS(c *config.Config, media model.Media, token string) (model.Video, error) {

	videoFile := &VideoFileReader{}

	// Create an instance of MultipartUploadHandlerHandlerInput
	input := MultipartUploadHandlerHandlerInput{
		Hostname:              fmt.Sprintf("%s:%s", c.APIConfig.URL, c.APIConfig.Port),
		Username:              c.APIConfig.Username,
		Password:              c.APIConfig.Password,
		ChannelID:             c.APIConfig.ChannelID, // replace with your channel ID
		File:                  videoFile,
		FileName:              media.FilePath,
		DisplayName:           media.Title,
		Privacy:               int8(c.APIConfig.Privacy), // replace with your privacy setting
		CommentsEnabled:       c.APIConfig.CommentsEnabled,
		DownloadEnabled:       c.APIConfig.DownloadEnabled,
		OriginallyPublishedAt: media.CreateDate.Format("2006-01-02 15:04:05"),
	}
	var err error
	input.File, err = GetVideoFileReader(input.FileName, VideoChunkSize)
	if err != nil {
		panic(err)
	}
	f, err := os.Open(input.FileName)
	if err != nil {

		logger.LogError("not able to open file", map[string]interface{}{"error": err, "file": input.FileName})
	}
	defer f.Close()
	input.ContentType, err = GetContentType(f)
	if err != nil {
		logger.LogError("not able to get content type .. skipping file", map[string]interface{}{"error": err, "file": input.FileName})
		return model.Video{}, err
	}

	// Call the function
	video, err := MultipartUploadHandler(input, token)

	if err != nil {
		logger.LogError("Error Uploading", map[string]interface{}{"error": err, "file": input.FileName})
		return model.Video{}, err
	}

	return video, nil
}
