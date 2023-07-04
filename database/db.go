package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"peertubeupload/config"

	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
)

func InitDB(c config.Config) (*sql.DB, error) {
	var connStr string
	var db *sql.DB
	var err error
	switch c.DBConfig.DBType {
	case "postgres":
		connStr = fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", c.DBConfig.Username, c.DBConfig.Password, c.DBConfig.Dbname, c.DBConfig.Host, c.DBConfig.Port)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, err
		}
		err = db.Ping()
		if err != nil {
			return nil, err
		}
		if !c.DBConfig.UpdateSameTable {
			err = checkAndCreateOrModifyPostgres(db, "peertube_log", c.DBConfig.ReferenceColumns...)
		}

		if err != nil {
			log.Println("Failed to check and create/modify table and columns:", err)
			os.Exit(1)
		}

	case "oracle":

		connStr := fmt.Sprintf("%s/%s@%s:%s/%s", c.DBConfig.Username, c.DBConfig.Password, c.DBConfig.Host, c.DBConfig.Port, c.DBConfig.Dbname)
		db, err := sql.Open("godror", connStr)
		if err != nil {
			return nil, err
		}
		err = db.Ping()
		if err != nil {
			return nil, err
		}
		if !c.DBConfig.UpdateSameTable {
			err = checkAndCreateOrModifyPostgres(db, "peertube_log", c.DBConfig.ReferenceColumns...)
		}

		if err != nil {
			log.Println("Failed to check and create/modify table and columns:", err)
			os.Exit(1)
		}

	}

	fmt.Println("Table and columns are checked and created/modified successfully!")

	return db, nil
}

// Function to check and create/modify table and columns for PostgreSQL
func checkAndCreateOrModifyPostgres(db *sql.DB, tableName string, columns ...string) error {
	// Check if the table exists
	tableExists := checkTableExists(db, tableName, "postgres")
	if !tableExists {
		// Create the table if it doesn't exist
		err := createTablePostgres(db, tableName, columns)
		if err != nil {
			return err
		}
	}

	// Check if the columns exist
	for _, column := range columns {
		columnExists, err := checkColumnExistsPostgres(db, tableName, column)
		if err != nil {
			return err
		}

		if !columnExists {
			// Add the column if it doesn't exist
			err := addColumnPostgres(db, tableName, column)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Function to check and create/modify table and columns for Oracle
func checkAndCreateOrModifyOracle(db *sql.DB, tableName string, columns ...string) error {
	// Check if the table exists
	tableExists := checkTableExists(db, tableName, "oracle")
	if !tableExists {
		// Create the table if it doesn't exist
		err := createTableOracle(db, tableName, columns)
		if err != nil {
			return err
		}
	}

	// Check if the columns exist
	for _, column := range columns {
		columnExists, err := checkColumnExistsOracle(db, tableName, column)
		if err != nil {
			return err
		}

		if !columnExists {
			// Add the column if it doesn't exist
			err := addColumnOracle(db, tableName, column)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Function to check if a table exists in the database for both PostgreSQL and Oracle
func checkTableExists(db *sql.DB, tableName string, dbType string) bool {
	var exists bool
	switch dbType {
	case "postgres":
		query := `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = $1
			)
		`
		row := db.QueryRow(query, tableName)
		err := row.Scan(&exists)
		if err != nil {
			return false
		}
	case "oracle":
		query := `
			SELECT 1
			FROM all_tables
			WHERE table_name = :1
		`
		stmt, err := db.Prepare(query)
		if err != nil {
			return false
		}
		defer stmt.Close()

		err = stmt.QueryRow(tableName).Scan(&exists)
		if err != nil {
			return false
		}
	}

	return exists
}

// Function to create the table for PostgreSQL
func createTablePostgres(db *sql.DB, tableName string, columns []string) error {
	columnDefinitions := ""
	for _, column := range columns {
		columnDefinitions += fmt.Sprintf("%s VARCHAR(255), ", column)
	}
	columnDefinitions = columnDefinitions[:len(columnDefinitions)-2] // Remove the trailing comma and space

	query := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, columnDefinitions)
	_, err := db.Exec(query)
	return err
}

// Function to create the table for Oracle
func createTableOracle(db *sql.DB, tableName string, columns []string) error {
	columnDefinitions := ""
	for _, column := range columns {
		columnDefinitions += fmt.Sprintf("%s VARCHAR2(255), ", column)
	}
	columnDefinitions = columnDefinitions[:len(columnDefinitions)-2] // Remove the trailing comma and space

	query := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, columnDefinitions)
	_, err := db.Exec(query)
	return err
}

// Function to check if a column exists in a table for PostgreSQL
func checkColumnExistsPostgres(db *sql.DB, tableName, columnName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			AND table_name = $1
			AND column_name = $2
		)
	`
	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(tableName, columnName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Function to check if a column exists in a table for Oracle
func checkColumnExistsOracle(db *sql.DB, tableName, columnName string) (bool, error) {
	query := `
		SELECT 1
		FROM all_tab_columns
		WHERE table_name = :1
		AND column_name = :2
	`
	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var exists int
	err = stmt.QueryRow(tableName, columnName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// Function to add a column to a table for PostgreSQL
func addColumnPostgres(db *sql.DB, tableName, columnName string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s VARCHAR(255)", tableName, columnName)
	_, err := db.Exec(query)
	return err
}

// Function to add a column to a table for Oracle
func addColumnOracle(db *sql.DB, tableName, columnName string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD %s VARCHAR2(255)", tableName, columnName)
	_, err := db.Exec(query)
	return err
}
