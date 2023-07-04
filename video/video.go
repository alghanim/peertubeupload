package video

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"peertubeupload/model"
)

func UploadVideo(baseURL string, client *http.Client, title string, description string, originalDateTime string, token string, filePath string) (model.Video, error) {
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
