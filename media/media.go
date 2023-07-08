package media

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"peertubeupload/config"
	"peertubeupload/logger"
	"peertubeupload/model"
)

func UploadMedia(baseURL string, client *http.Client, title string, description string, originalDateTime string, token string, fPath string, c *config.Config) (model.Video, error) {
	url := baseURL + "/videos/upload"
	method := "POST"

	var tmpfilePath string
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	metadata, err := getMetaData(fPath)
	if err != nil {
		return model.Video{}, err
	}
	// Check if the file is audio or video
	isAudio := false
	isVideo := false

	for _, stream := range metadata.Streams {
		if stream.CodecType == "video" {
			isVideo = true

		}
		if stream.CodecType == "audio" {
			isAudio = true
		}
	}
	if isVideo {
		file, errFile1 := os.Open(fPath)
		if errFile1 != nil {
			return model.Video{}, errFile1
		}
		defer file.Close()
		part1, errFile1 := writer.CreateFormFile("videofile", filepath.Base(fPath))
		if errFile1 != nil {
			return model.Video{}, errFile1
		}
		_, errFile1 = io.Copy(part1, file)
		if errFile1 != nil {

			return model.Video{}, errFile1
		}

	} else if isAudio {

		filename := GetFileName(fPath)

		if c.LoadType.ConvertAudioToMp3 {
			tmpfilePath = fmt.Sprintf("%s.mp3", path.Join(c.LoadType.TempFolder, filename))

			err := convertToMp3(fPath, tmpfilePath)
			if err != nil {
				return model.Video{}, err
			}
			fPath = tmpfilePath
		}

		file, errFile1 := os.Open(fPath)
		if errFile1 != nil {
			return model.Video{}, errFile1
		}
		defer file.Close()
		part1, errFile1 := writer.CreateFormFile("videofile", filepath.Base(fPath))
		if errFile1 != nil {
			return model.Video{}, errFile1
		}
		_, errFile1 = io.Copy(part1, file)
		if errFile1 != nil {

			return model.Video{}, errFile1
		}

	}

	// _ = writer.WriteField("channelId", c.APIConfig.ChannelID)
	// _ = writer.WriteField("downloadEnabled", c.APIConfig.DownloadEnabled)
	// _ = writer.WriteField("name", title)
	// // _ = writer.WriteField("description", description)
	// _ = writer.WriteField("commentsEnabled", c.APIConfig.CommentsEnabled)
	// _ = writer.WriteField("originallyPublishedAt", originalDateTime)
	// _ = writer.WriteField("privacy", c.APIConfig.Privacy)
	// _ = writer.WriteField("waitTranscoding", c.APIConfig.WaitTranscoding)
	err = writer.Close()
	if err != nil {

		return model.Video{}, err
	}

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
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

	os.Remove(tmpfilePath)
	return video, nil
}

func getMetaData(filepath string) (model.Metadata, error) {

	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filepath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		logger.LogError("ffprobe didn't work properly", map[string]interface{}{"error": err, "function": "getMetaData"})
		return model.Metadata{}, err
	}

	metadata, err := model.UnmarshalMetadata(output)
	if err != nil {
		logger.LogError("metadata ubnarshal is not done", map[string]interface{}{"error": err, "function": "getMetaData"})
		return model.Metadata{}, err
	}

	return metadata, nil
}

func GetFileName(fpath string) string {

	filenameWithExt := filepath.Base(fpath) // get the file name with extension

	filename := filenameWithExt[0 : len(filenameWithExt)-len(filepath.Ext(fpath))] // remove the extension

	return filename
}

func convertToMp3(fpath string, tmpPath string) error {

	cmd := exec.Command("ffmpeg", "-y", "-i", fpath, tmpPath)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
