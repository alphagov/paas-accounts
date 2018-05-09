package db

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

func sqlDir() string {
	root := os.Getenv("APP_ROOT")
	if root == "" {
		root = os.Getenv("PWD")
	}
	if root == "" {
		root, _ = os.Getwd()
	}
	return filepath.Join(root, "db", "sql")
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
