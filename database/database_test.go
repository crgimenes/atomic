package database

import (
	"testing"

	_ "modernc.org/sqlite"
)

func TestNew(t *testing.T) {
	connectionString = ":memory:"

	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		e := db.Close()
		if e != nil {
			t.Fatal(e)
		}
	}()

	version, err := db.GetVersion()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(version)
}

func TestDatabase_RunMigration(t *testing.T) {
	connectionString = ":memory:"

	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		e := db.Close()
		if e != nil {
			t.Fatal(e)
		}
	}()

	err = db.createMigrationTable()
	if err != nil {
		t.Fatal(err)
	}

	m, err := db.VerifyMigration()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Migration:", m)

	err = db.RunMigration()
	if err != nil {
		t.Fatal(err)
	}

	m, err = db.VerifyMigration()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Migration:", m)

	err = db.RunMigration()
	if err != nil {
		t.Fatal(err)
	}

	if m != currentMigration {
		t.Fatal("Migration not up to date")
	}
}

func TestDatabase_ChkMigration(t *testing.T) {
	connectionString = ":memory:"

	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		e := db.Close()
		if e != nil {
			t.Fatal(e)
		}
	}()

	err = db.createMigrationTable()
	if err != nil {
		t.Fatal(err)
	}

	err = db.ChkMigration()
	if err != ErrDatabaseNotUpToDate {
		t.Fatal("Expected ErrDatabaseNotUpToDate")
	}

	err = db.RunMigration()
	if err != nil {
		t.Fatal(err)
	}

	err = db.ChkMigration()
	if err != nil {
		t.Fatal(err)
	}

	// insert a fake migration
	_, err = db.db.Exec("INSERT INTO migrations (id) VALUES (999999)")
	if err != nil {
		t.Fatal(err)
	}

	err = db.ChkMigration()
	if err != ErrDatabaseAhead {
		t.Fatal("Expected ErrDatabaseAhead")
	}

}

func TestDatabase_CreateUser(t *testing.T) {
	connectionString = ":memory:"

	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		e := db.Close()
		if e != nil {
			t.Fatal(e)
		}
	}()

	err = db.createMigrationTable()
	if err != nil {
		t.Fatal(err)
	}

	err = db.RunMigration()
	if err != nil {
		t.Fatal(err)
	}

	nickname := "test"
	email := "test@test"
	password := "test1234567"
	sshPublicKey := ""
	groups := "group1,group2"
	u, err := db.CreateUser(
		nickname,
		email,
		password,
		sshPublicKey,
		groups,
	)
	if err != nil {
		t.Fatal(err)
	}

	if u.Email != email {
		t.Fatal("Email not equal")
	}

	if u.Password == password {
		t.Fatal("Password not encrypted")
	}

	u, err = db.CreateUser(
		"",
		email,
		password,
		sshPublicKey,
		groups,
	)
	if err != ErrNicknameEmpty {
		t.Fatal("Expected ErrNicknameEmpty")
	}

	u, err = db.CreateUser(
		nickname,
		"",
		password,
		sshPublicKey,
		groups,
	)
	if err != ErrEmailEmpty {
		t.Fatal("Expected ErrEmailEmpty")
	}

	u, err = db.CreateUser(
		nickname,
		email,
		"",
		"",
		groups,
	)
	if err != ErrPasswordOrSSHKeyRequired {
		t.Fatal("Expected ErrPasswordOrSSHKeyRequired")
	}

	nickname = "test2"
	email = "test2@test"
	password = "test1234567"

	u, err = db.CreateUser(
		nickname,
		email,
		password,
		sshPublicKey,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if u.Groups != "users" {
		t.Fatal("Expected groups to be users")
	}

	u, err = db.CreateUser(
		nickname,
		email,
		"1234",
		sshPublicKey,
		groups,
	)
	if err != ErrPasswordTooShort {
		t.Fatal("Expected ErrPasswordTooShort")
	}

	u, err = db.GetUserByID(2)
	if err != nil {
		t.Fatal(err)
	}

	if u.Nickname != nickname {
		t.Fatal("Nickname not equal")
	}

	u, err = db.GetUserByNickname(nickname)
	if err != nil {
		t.Fatal(err)
	}

	if u.Nickname != nickname {
		t.Fatal("Nickname not equal")
	}

	u, err = db.GetUserByEmail(email)
	if err != nil {
		t.Fatal(err)
	}

	if u.Email != email {
		t.Fatal("Email not equal")
	}

}

func TestDatabase_CheckAndReturnUser(t *testing.T) {

	connectionString = ":memory:"

	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		e := db.Close()
		if e != nil {
			t.Fatal(e)
		}
	}()

	err = db.createMigrationTable()
	if err != nil {
		t.Fatal(err)
	}

	err = db.RunMigration()
	if err != nil {
		t.Fatal(err)
	}

	nickname := "test"
	email := "test@test"
	password := "test1234567"

	u, err := db.CreateUser(
		nickname,
		email,
		password,
		"",
		"",
	)
	if err != nil {
		t.Fatal(err)
	}

	u, err = db.CheckAndReturnUser(nickname, password)
	if err != nil {
		t.Fatal(err)
	}

	if u.Nickname != nickname {
		t.Fatal("Nickname not equal")
	}

	u, err = db.CheckAndReturnUser(nickname, "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Fatal("Expected ErrInvalidCredentials")
	}

	u, err = db.CheckAndReturnUser("wrongnickname", password)
	if err != ErrInvalidCredentials {
		t.Fatal("Expected ErrInvalidCredentials")
	}

}
