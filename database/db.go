package database

import (
	"database/sql"
	"fmt"
	"peertubeupload/config"

	_ "github.com/lib/pq"
)

func InitDB(c config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", c.DBConfig.Username, c.DBConfig.Password, c.DBConfig.Dbname, c.DBConfig.Host, c.DBConfig.Port)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
