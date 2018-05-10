package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

type User struct {
	UUID string
}

type Document struct {
	Name, Content string
	ValidFrom     time.Time
}

type Agreement struct {
	UserUUID     string
	DocumentName string
	Date         time.Time
}

func sqlDir() string {
	root := os.Getenv("APP_ROOT")
	if root == "" {
		root = os.Getenv("PWD")
	}
	if root == "" {
		root, _ = os.Getwd()
	}
	return filepath.Join(root, "database", "sql")
}

type DB struct {
	conn *sql.DB
}

func NewDB(connstr string) (*DB, error) {
	conn, err := sql.Open("postgres", connstr)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Init() error {
	driver, err := postgres.WithInstance(db.conn, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance("file://"+sqlDir(), "postgres", driver)
	if err != nil {
		return err
	}

	defer m.Close()
	if err := m.Up(); err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (db *DB) PutDocument(doc Document) error {
	_, err := db.conn.Exec(`INSERT INTO documents (name, content, valid_from) VALUES ($1, $2, $3)`, doc.Name, doc.Content, doc.ValidFrom)

	return err
}

func (db *DB) GetDocument(name string) (Document, error) {
	doc := Document{}
	err := db.conn.QueryRow(`SELECT name, content, valid_from FROM documents WHERE name = $1 ORDER BY valid_from DESC LIMIT 1`, name).Scan(&doc.Name, &doc.Content, &doc.ValidFrom)

	return doc, err
}

func (db *DB) PutUser(user User) error {
	_, err := db.conn.Exec(`INSERT INTO users (uuid) VALUES ($1) ON CONFLICT DO NOTHING`, user.UUID)

	return err
}

func (db *DB) PutAgreement(agreement Agreement) error {
	_, err := db.conn.Exec(`
		INSERT INTO agreements (
			user_uuid, document_name, date
		) VALUES (
			$1, $2, $3
		)
	`, agreement.UserUUID, agreement.DocumentName, agreement.Date)

	return err
}

func (db *DB) GetAgreementsForUserUUID(uuid string) ([]Agreement, error) {
	rows, err := db.conn.Query(`
		SELECT
			user_uuid, document_name, date
		FROM
			agreements
		WHERE
			user_uuid = $1
		ORDER BY
			date
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	agreements := []Agreement{}
	for rows.Next() {
		var agreement Agreement
		err := rows.Scan(&agreement.UserUUID, &agreement.DocumentName, &agreement.Date)
		if err != nil {
			return nil, err
		}
		agreements = append(agreements, agreement)
	}

	return agreements, nil
}
