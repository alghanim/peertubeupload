package medialog

import (
	"database/sql"
	"encoding/json"
	"os"
	"peertubeupload/config"
	"peertubeupload/model"
)

func LogResultToFile(media model.Video, f model.Media, c *config.Config) error {

	// Open the file in append mode
	file, err := os.OpenFile("log.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	defer file.Close()
	combined := struct {
		Media model.Media
		Video model.Video
	}{
		Media: f,
		Video: media,
	}

	err = encoder.Encode(combined)
	if err != nil {

		return err
	}
	return nil
}
func LogResultToDB(media model.Video, f model.Media, c *config.Config, db *sql.DB) error {

	switch c.DBConfig.DBType {
	case "postgres":
		insertQuery := `
			INSERT INTO public.peertube_log (id, uuid, shortuuid, file_path)
		VALUES ($1, $2, $3, $4)
		`
		_, err := db.Exec(insertQuery, media.Video.ID, media.Video.UUID, media.Video.ShortUUID, f.FilePath)
		if err != nil {
			return err
		}
	case "oracle":
		insertQuery := `
			INSERT INTO peertube_log (id, uuid, shortuuid, file_path)
		VALUES (:1, :2, :3, :4)
		`
		_, err := db.Exec(insertQuery, media.Video.ID, media.Video.UUID, media.Video.ShortUUID, f.FilePath)
		if err != nil {
			return err
		}

	}

	return nil
}
