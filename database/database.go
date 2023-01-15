package database

import (
	"database/sql"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const (
	// currentMigration is the current migration version of the code.
	// it must be incremented every time a new migration is added.
	currentMigration = 1
)

var (
	ErrDatabaseAhead            = errors.New("database is ahead of current migration, please update the application")
	ErrDatabaseNotUpToDate      = errors.New("database is not up to date, please run migrations")
	ErrNicknameEmpty            = errors.New("nickname is required")
	ErrEmailEmpty               = errors.New("email is required")
	ErrPasswordOrSSHKeyRequired = errors.New("password or ssh public key is required")
	ErrPasswordTooShort         = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials       = errors.New("invalid credentials")

	connectionString = `file:atomic.db?mode=rwc&_journal_mode=WAL&_busy_timeout=10000`

	// go:embed migration01.sql
	migration01 string
)

type Database struct {
	db *sqlx.DB
}

func New() (*Database, error) {
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
	Groups       string `db:"groups"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

func (d *Database) RunMigration() error {
	err := d.createMigrationTable()
	if err != nil {
		return err
	}

	lastMigration, err := d.VerifyMigration()
	if err != nil {
		return err
	}

	log.Printf("last migration: %d", lastMigration)

	// begin transaction
	tx, err := d.db.Beginx()
	// run migrations
	switch lastMigration {
	case 0:
		log.Println("running migration 1")
		_, err = tx.Exec(migration01)
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

func (d *Database) createMigrationTable() error {
	_, err := d.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
				id INTEGER PRIMARY KEY,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`)
	return err
}

func (d *Database) VerifyMigration() (int, error) {
	var (
		lastMigration int
		count         int
	)
	err := d.db.Get(&count, `SELECT COUNT(*) FROM migrations`)
	if err != nil {
		return 0, err
	}
	log.Printf("migrations: %d", count)
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

func (d *Database) CreateUser(nickname, email, password, sshPublicKey, groups string) (User, error) {

	if nickname == "" {
		return User{}, ErrNicknameEmpty
	}
	if email == "" {
		return User{}, ErrEmailEmpty
	}
	if password == "" && sshPublicKey == "" {
		return User{}, ErrPasswordOrSSHKeyRequired
	}
	if password != "" {
		if len(password) < 8 {
			return User{}, ErrPasswordTooShort
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		password = string(hashedPassword)
	}
	if groups == "" {
		groups = "users"
	}

	sql := `INSERT INTO users (
		nickname,
		email,
		password,
		ssh_public_key, 
		groups) 
		VALUES ($1, $2, $3, $4, $5) RETURNING *`
	var user User
	err := d.db.QueryRowx(sql, nickname, email, password, sshPublicKey, groups).StructScan(&user)

	return user, err
}

func (d *Database) GetUserByID(id int) (User, error) {
	var user User
	err := d.db.QueryRowx(`SELECT * FROM users WHERE id = $1`, id).StructScan(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (d *Database) GetUserByEmail(email string) (User, error) {
	var user User
	err := d.db.QueryRowx(`SELECT * FROM users WHERE email = $1`, email).StructScan(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (d *Database) GetUserByNickname(nickname string) (User, error) {
	var user User
	err := d.db.QueryRowx(`SELECT * FROM users WHERE nickname = $1`, nickname).StructScan(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (d *Database) CheckAndReturnUser(nickname, password string) (User, error) {
	user, err := d.GetUserByNickname(nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	if user.Password == "" {
		return User{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}
