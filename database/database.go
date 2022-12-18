package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"crg.eti.br/go/atomic/config"
)

type Database struct {
	db *sqlx.DB
}

func New(cfg config.Config) (*Database, error) {
	connectionString := fmt.Sprintf(`file:%s?mode=rwc&_journal_mode=WAL&_busy_timeout=10000`, cfg.DBFile)
	db, err := sqlx.Open("sqlite", connectionString)
	if err != nil {
		return nil, err
	}
	return &Database{
		db: db,
	}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetVersion() (string, error) {
	var version string
	err := d.db.Get(&version, "SELECT sqlite_version()")
	if err != nil {
		return "", err
	}
	return version, nil
}

type User struct {
	ID           int    `db:"id"`
	Nickname     string `db:"nickname"`
	Email        string `db:"email"`
	Password     string `db:"password"`
	SSHPublicKey string `db:"ssh_public_key"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

func (d *Database) GetUserByName(name string) (User, error) {
	var user User
	err := d.db.QueryRowx(`SELECT * FROM users WHERE nickname = ?`, name).StructScan(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
