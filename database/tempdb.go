package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

type TempDB struct {
	MasterConnectionString string
	TempConnectionString   string
}

func NewTempDB() (*TempDB, error) {
	masterConnectionString := os.Getenv("TEST_DATABASE_URL")
	if masterConnectionString == "" {
		masterConnectionString = "postgres://postgres:postgres@localhost:5432/?sslmode=disable"
	}
	master, err := sql.Open("postgres", masterConnectionString)
	if err != nil {
		return nil, err
	}
	defer master.Close()
	dbName := "test_" + strings.Replace(uuid.NewV4().String(), "-", "_", -1)
	_, err = master.Exec(fmt.Sprintf(`CREATE DATABASE %s`, dbName))
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(masterConnectionString)
	if err != nil {
		return nil, err
	}
	u.Path = "/" + dbName
	tdb := &TempDB{
		TempConnectionString:   u.String(),
		MasterConnectionString: masterConnectionString,
	}
	return tdb, nil
}

// Close drops the database
func (db *TempDB) Close() error {
	conn, err := sql.Open("postgres", db.MasterConnectionString)
	if err != nil {
		return err
	}
	defer conn.Close()
	u, err := url.Parse(db.TempConnectionString)
	if err != nil {
		return err
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	try := 0
	for {
		_, err = conn.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, dbName))
		if err != nil {
			if try > 3 {
				return err
			}
			fmt.Println(err)
			try++
			time.Sleep(1 * time.Second)
			continue
		}
		return nil
	}
}
