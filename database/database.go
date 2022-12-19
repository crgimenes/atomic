package database

import (
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const (
	// currentMigration is the current migration version of the code.
	// it must be incremented every time a new migration is added.
	currentMigration = 1
)

var (
	ErrDatabaseAhead       = errors.New("database is ahead of current migration, please update the application")
	ErrDatabaseNotUpToDate = errors.New("database is not up to date, please run migrations")
)

type Database struct {
	db *sqlx.DB
}

func New() (*Database, error) {
	connectionString := `file:atomic.db?mode=rwc&_journal_mode=WAL&_busy_timeout=10000`
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

func (d *Database) RunMigration() error {
	// migration table
	_, err := d.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
				id INTEGER PRIMARY KEY,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`)
	if err != nil {
		return err
	}

	lastMigration, err := d.VerifyMigration()
	if err != nil {
		return err
	}

	log.Printf("Last migration: %d", lastMigration)

	// begin transaction
	tx, err := d.db.Beginx()
	// run migrations
	switch lastMigration {
	case 0:
		log.Println("running migration 1")
		_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS users (
					id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
					nickname TEXT NOT NULL,
					email TEXT NOT NULL,
					password TEXT NOT NULL,
					ssh_public_key TEXT NOT NULL,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		// update migration table
		_, err = tx.Exec(`INSERT INTO migrations (id) VALUES (1)`)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		log.Println("done migration 1")
		lastMigration = 1

		fallthrough
	default:
		log.Println("no migrations to run")
	}

	if currentMigration != lastMigration {
		_ = tx.Rollback()

		// this should never happen... ok it can happen if you forget
		// to update the currentMigration variable.
		log.Fatal("currentMigration variable is not up to date")
	}

	return tx.Commit()
}

func (d *Database) VerifyMigration() (int, error) {
	var lastMigration = 0
	// count migrations
	var count int
	err := d.db.Get(&count, `SELECT COUNT(*) FROM migrations`)
	if err != nil {
		return 0, err
	}
	log.Printf("Migrations: %d", count)
	if count != 0 {
		err = d.db.Get(&lastMigration, "SELECT MAX(id) as max FROM migrations")
		if err != nil {
			return 0, err
		}
	}

	return lastMigration, nil
}

func (d *Database) ChkMigration() error {
	lastMigration, err := d.VerifyMigration()
	if err != nil {
		return err
	}

	if lastMigration < currentMigration {
		return ErrDatabaseNotUpToDate
	}

	if lastMigration > currentMigration {
		return ErrDatabaseAhead
	}

	return nil
}
