package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PeertubeUserToken struct {
	Access_token string
}
type OauthClientsLocal struct {
	Client_id     string
	Client_secret string
}
type MultipartUploadHandlerHandlerInput struct {
	Hostname        string
	Username        string
	Password        string
	ContentType     string
	ChannelID       int
	File            *VideoFileReader
	FileName        string
	DisplayName     string
	Privacy         int8
	Category        int
	CommentsEnabled bool
	DescriptionText string
	DownloadEnabled bool
	Language        string
	Licence         int
	NSFW            bool
	SupportText     string
	Tags            []string
}

func GetClientLocal(hostname string) (*OauthClientsLocal, error) {
	/*
		Get oauth local client via
		https://docs.joinpeertube.org/api-rest-getting-started?id=get-client
	*/
	clientslocalurl := fmt.Sprintf("%s/api/v1/oauth-clients/local", hostname)
	clientslocalreq, err := http.Get(clientslocalurl)
	if err != nil {
		return nil, err
	}
	defer clientslocalreq.Body.Close()
	if clientslocalreq.StatusCode != 200 {
		switch clientslocalreq.StatusCode {
		case 423:
			return nil, fmt.Errorf("too many requests")
		default:
			return nil, errors.New(fmt.Sprintf("clientslocalreq url %s returned status code %d.", clientslocalurl, clientslocalreq.StatusCode))
		}
	}
	body, err := io.ReadAll(clientslocalreq.Body)
	if err != nil {
		return nil, err
	}

	local := new(OauthClientsLocal)

	if err = json.Unmarshal(body, local); err != nil {
		return nil, err
	}

	return local, nil
}
func GetUserTokenFromAPI(hostname, username, password string) (string, error) {
	clientlocal, err := GetClientLocal(hostname)
	if err != nil {
		return "", err
	}

	postForm := url.Values{}
	postForm.Add("client_id", clientlocal.Client_id)
	postForm.Add("client_secret", clientlocal.Client_secret)
	postForm.Add("grant_type", "password")
	postForm.Add("response_type", "code")
	postForm.Add("username", username)
	postForm.Add("password", password)

	gutUrl := fmt.Sprintf("%s/api/v1/users/token", hostname)
	tokenreq, err := http.PostForm(gutUrl, postForm)
	if err != nil {
		return "", err
	}
	defer tokenreq.Body.Close()

	if tokenreq.StatusCode != 200 {
	}

	tokenreqbody, err := io.ReadAll(tokenreq.Body)
	if err != nil {
		return "", err
	}
	tokenStr := new(PeertubeUserToken)
	if err = json.Unmarshal(tokenreqbody, tokenStr); err != nil {
		return "", err
	}
	return tokenStr.Access_token, nil
}

func MultipartUploadHandler(input MultipartUploadHandlerHandlerInput) (err error) {
	oauthToken, err := GetUserTokenFromAPI(input.Hostname, input.Username, input.Password)
	if err != nil {
		return
	}

	client := &http.Client{}
	initializeUrl := fmt.Sprintf("%s/api/v1/videos/upload-resumable", input.Hostname)
	initializePayload := map[string]interface{}{
		"channelId": input.ChannelID,
		"filename":  input.FileName,
		"name":      input.DisplayName,
	}
	initializePayloadBytes, err := json.Marshal(initializePayload)
	if err != nil {
		panic(err)
	}
	initialize, err := http.NewRequest("POST", initializeUrl, bytes.NewReader(initializePayloadBytes))
	if err != nil {
		panic(err)
	}

	initialize.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oauthToken))
	initialize.Header.Add("X-Upload-Content-Length", fmt.Sprintf("%d", input.File.TotalBytes))
	initialize.Header.Add("X-Upload-Content-Type", input.ContentType)
	initialize.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(initialize)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 201 {
		log.Printf("initialize api call returned status code %d.", resp.StatusCode)
		fmt.Println(resp.Header)
		body, err2 := io.ReadAll(resp.Body)
		if err2 != nil {
			panic(err)
		}
		defer resp.Body.Close()
		fmt.Println(string(body))
		panic("returned non 201 status code")
	}
	fmt.Printf("%+v\n", resp.Header)

	defer resp.Body.Close()
	uploadLocation := resp.Header.Get("Location")

	if strings.HasPrefix(uploadLocation, "//") {
		uploadLocation = "http:" + uploadLocation
	} else {
		log.Println("Warning: recieved an upload location that doesn't begin with \"//\", i don't know what to do with this.")
		panic(nil)
	}
	fmt.Println("upload location", uploadLocation)

	for {
		chunk, err := input.File.GetNextChunk()
		if err != nil {
			panic(err)
		}
		if chunk.Finished {
			break
		}
		fmt.Println(chunk.MinByte, chunk.MaxByte, chunk.Length, chunk.RangeHeader)

		for {
			up, err := http.NewRequest("PUT", uploadLocation, bytes.NewReader(chunk.Bytes))
			if err != nil {
				panic(err)
			}

			up.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oauthToken))
			up.Header.Add("Content-Length", fmt.Sprintf("%d", chunk.Length))
			up.Header.Add("Content-Range", chunk.RangeHeader)

			resp, err := client.Do(up)
			if err != nil {
				panic(err)
			}
			fmt.Println(resp.Status)
			defer resp.Body.Close()
			body, err2 := io.ReadAll(resp.Body)
			if err2 != nil {
				panic(err2)
			}
			fmt.Println(string(body))
			if resp.StatusCode == 308 || resp.StatusCode == 200 {
				break
			} else {
				log.Println("Status code other than 308 or 200 recieved. Will retry.")
				time.Sleep(15 * time.Second)
			}
		}
	}
	return
}
func test() {
	// set video file

	// Create an instance of VideoFileReader
	// This is just a placeholder. You will need to replace it with actual code.
	videoFile := &VideoFileReader{}

	// Create an instance of MultipartUploadHandlerHandlerInput
	input := MultipartUploadHandlerHandlerInput{
		Hostname:        "http://peertube.alghanim:9000",
		Username:        "root",
		Password:        "ali12345",
		ContentType:     "video/mp4",
		ChannelID:       1, // replace with your channel ID
		File:            videoFile,
		FileName:        "/mnt/c/Users/algha/Videos/jelly/The.Adam.Project.2022.1080p.WEBRip.x264.AAC5.1-[YTS.MX].mp4",
		DisplayName:     "your-video-display-name",
		Privacy:         1, // replace with your privacy setting
		Category:        1, // replace with your category
		CommentsEnabled: true,
		DescriptionText: "your-video-description",
		DownloadEnabled: true,
		Language:        "en", // replace with your language
		Licence:         1,    // replace with your licence
		NSFW:            false,
		SupportText:     "your-support-text",
		Tags:            []string{"tag1", "tag2"}, // replace with your tags
	}
	var err error
	input.File, err = GetVideoFileReader(input.FileName, VideoChunkSize)
	if err != nil {
		panic(err)
	}

	// Call the function
	err = MultipartUploadHandler(input)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
}
