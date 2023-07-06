package medialog

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"peertubeupload/config"
	"peertubeupload/model"
	"strings"
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
func LogResultToDB(media model.Video, f map[string]interface{}, c *config.Config, db *sql.DB, fPath string) error {
	var logTableName string
	if c.LoadType.LoadPathFromDB {
		logTableName = fmt.Sprintf("%s_to_peertube_log", c.DBConfig.TableName)
	} else {
		logTableName = "peertube_log"
	}

	combinedColumns := append(c.DBConfig.ReferenceColumns, c.DBConfig.MediaIdentifier...)

	shit, err := mergeStructAndMap(media.Video, f)
	if err != nil {
		return err
	}
	// Create a slice to hold the values to be inserted
	values := make([]interface{}, len(combinedColumns))
	for i, column := range combinedColumns {
		value, ok := shit[strings.ToLower(column)]
		if !ok {
			return fmt.Errorf("column %s not found in map", column)
		}
		values[i] = value
	}

	switch c.DBConfig.DBType {
	case "postgres":
		// Create the placeholders for the values
		params := make([]string, len(combinedColumns))
		for i := range params {
			params[i] = fmt.Sprintf("$%d", i+1)
		}
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			logTableName,
			strings.Join(combinedColumns, ", "),
			strings.Join(params, ", "),
		)
		_, err := db.Exec(insertQuery, values...)
		if err != nil {
			return err
		}
	case "oracle":
		// Create the placeholders for the values
		params := make([]string, len(combinedColumns))
		for i := range params {
			params[i] = fmt.Sprintf(":%d", i+1)
		}
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			logTableName,
			strings.Join(combinedColumns, ", "),
			strings.Join(params, ", "),
		)
		_, err := db.Exec(insertQuery, values...)
		if err != nil {
			return err
		}
	}

	return nil
}

func mergeStructAndMap(s model.VideoClass, m map[string]interface{}) (map[string]interface{}, error) {

	ptid := struct {
		Peertubeid int64  `json:"peertube_id"`
		UUID       string `json:"uuid"`
		ShortUUID  string `json:"shortuuid"`
	}{
		Peertubeid: s.ID,
		UUID:       s.UUID,
		ShortUUID:  s.ShortUUID,
	}
	// First, marshal the struct to JSON
	jsonBytes, err := json.Marshal(ptid)
	if err != nil {
		return nil, err
	}

	// Then, unmarshal the JSON into a map
	var smap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &smap)
	if err != nil {
		return nil, err
	}

	// Now, add the keys and values from 'm' to 'smap'
	for k, v := range m {

		smap[k] = v

	}

	return smap, nil
}
