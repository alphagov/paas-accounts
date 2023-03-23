package database

import (
	"database/sql"
	"embed"
	"errors"
	"strings"
	"time"

	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/lib/pq"
)

type User struct {
	UUID     string  `json:"user_uuid" validate:"uuid"`
	Email    *string `json:"user_email" validate:"omitempty,email"`
	Username *string `json:"username" validate:"required,min=1"`
}

type Document struct {
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	ValidFrom time.Time `json:"valid_from"`
}

type Agreement struct {
	UserUUID     string    `json:"user_uuid"`
	DocumentName string    `json:"document_name"`
	Date         time.Time `json:"date"`
}

type UserDocument struct {
	Name          string     `json:"name"`
	Content       string     `json:"content"`
	ValidFrom     time.Time  `json:"valid_from"`
	AgreementDate *time.Time `json:"agreement_date"`
}

var (
	//go:embed sql/*.sql
	sqlFs embed.FS

	ErrDocumentNotFound = errors.New("document not found")
	ErrUserNotFound     = errors.New("user not found")
)

type DB struct {
	conn    *sql.DB
	connstr string
}

func NewDB(connstr string) (*DB, error) {
	conn, err := sql.Open("postgres", connstr)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn, connstr: connstr}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Init() error {
	sourceDriver, err := iofs.New(sqlFs, "sql")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, db.connstr)
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
	latestDocVersion, err := db.GetDocument(doc.Name)
	if err != nil && err != ErrDocumentNotFound {
		return err
	}

	if err == ErrDocumentNotFound || latestDocVersion.Content != doc.Content {
		_, err = db.conn.Exec(`INSERT INTO documents (name, content, valid_from) VALUES ($1, $2, $3)`, doc.Name, doc.Content, doc.ValidFrom)
		return err
	}

	return nil
}

func (db *DB) GetDocument(name string) (Document, error) {
	doc := Document{}
	err := db.conn.QueryRow(`SELECT name, content, valid_from FROM documents WHERE name = $1 ORDER BY valid_from DESC LIMIT 1`, name).Scan(&doc.Name, &doc.Content, &doc.ValidFrom)

	if err == sql.ErrNoRows {
		err = ErrDocumentNotFound
	}

	return doc, err
}

func (db *DB) PostUser(user User) error {
	_, err := db.GetUser(user.UUID)
	if err == ErrUserNotFound {
		_, err = db.conn.Exec(`INSERT INTO users (uuid, email, username) VALUES ($1, $2, $3)`, user.UUID, lowerStrPoint(user.Email), user.Username)
	}
	return err
}

func (db *DB) PatchUser(user User) error {
	_, err := db.conn.Exec(`UPDATE users SET email = $2, username = $3 WHERE uuid = $1`, user.UUID, lowerStrPoint(user.Email), user.Username)
	return err
}

func (db *DB) GetUser(uuid string) (User, error) {
	user := User{}
	err := db.conn.QueryRow(`
		SELECT uuid, email, username FROM users WHERE uuid = $1
	`, uuid).Scan(&user.UUID, &user.Email, &user.Username)

	if err == sql.ErrNoRows {
		err = ErrUserNotFound
	}

	return user, err
}

func (db *DB) GetUserByEmail(email string) ([]*User, error) {
	var users []*User
	rows, err := db.conn.Query(`
		SELECT uuid, email, username FROM users WHERE email = $1
	`, email)

	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return users, nil
		}

		return users, err
	}

	for rows.Next() {
		var user User
		err := rows.Scan(&user.UUID, &user.Email, &user.Username)
		if err != nil {
			return users, err
		}

		users = append(users, &user)
	}

	return users, err
}

func (db *DB) GetUserByUsername(username string) (User, error) {
	user := User{}
	err := db.conn.QueryRow(`
		SELECT uuid, email, username FROM users WHERE username = $1
	`, username).Scan(&user.UUID, &user.Email, &user.Username)

	if err == sql.ErrNoRows {
		err = ErrUserNotFound
	}

	return user, err
}

func (db *DB) GetUsersByUUID(uuids []string) ([]*User, error) {

	users := []*User{}

	if len(uuids) == 0 {
		return users, nil
	}

	uuidsCopy := make([]interface{}, len(uuids))
	for i, v := range uuids {
		uuidsCopy[i] = v
	}

	var f strings.Builder
	for i := range uuids {
		f.WriteString(fmt.Sprintf("$%v,", i+1))
	}
	fragment := strings.TrimSuffix(f.String(), ",")
	query := strings.Replace(`SELECT uuid, email, username FROM users WHERE uuid IN (uuids)`, "uuids", fragment, -1)

	rows, err := db.conn.Query(query, uuidsCopy...)
	if err != nil {
		return users, err
	}

	defer rows.Close()

	// Map UUID strings to user instances
	uuidToUser := map[string]*User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.UUID, &user.Email, &user.Username)
		if err != nil {
			return users, err
		}
		uuidToUser[user.UUID] = &user
	}

	// Results should be in the same order as input
	// and if a result wasn't found for an input it
	// should be mapped to nil
	for _, uuid := range uuids {
		if usr, ok := uuidToUser[uuid]; ok {
			users = append(users, usr)
		} else {
			users = append(users, nil)
		}
	}

	return users, nil
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

func (db *DB) GetDocumentsForUserUUID(uuid string) ([]UserDocument, error) {
	rows, err := db.conn.Query(`
		WITH valid_documents AS (
			SELECT
				*,
				tstzrange(valid_from, lead(valid_from, 1, 'infinity') over (
						partition by name order by valid_from rows between current row and 1 following
				)) as valid_for
			FROM
				documents
		)
		SELECT
			d.name,
			d.content,
			d.valid_from,
			agreements.date
		FROM
			valid_documents d
		LEFT JOIN
			agreements ON (
				d.name = agreements.document_name
				AND agreements.date <@ d.valid_for
				AND agreements.user_uuid = $1
			)
		ORDER BY
			agreements.date
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userDocuments := []UserDocument{}
	for rows.Next() {
		var userDocument UserDocument
		var nullTime pq.NullTime
		err := rows.Scan(&userDocument.Name, &userDocument.Content, &userDocument.ValidFrom, &nullTime)
		if err != nil {
			return nil, err
		}
		if nullTime.Valid {
			userDocument.AgreementDate = &nullTime.Time
		}
		userDocuments = append(userDocuments, userDocument)
	}
	return userDocuments, nil
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

func (db *DB) Ping() error {
	return db.conn.Ping()
}

func lowerStrPoint(str *string) *string {
	if str == nil {
		return nil
	}

	lower := strings.ToLower(*str)
	return &lower
}
