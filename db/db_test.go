package db_test

import (
	"database/sql"
	sqlDb "myproject/db"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const dbFile = "test.db"

var (
	db *sql.DB
)

func setup() {
	_db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}

	statment := `
	create table if not exists user (
		id integer not null primary key, 
		username text, 
		email text, 
		password text, 
		session text
	);
	`

	if _, err := _db.Exec(statment); err != nil {
		panic(err)
	}

	db = _db
}

func TestMain(m *testing.M) {
	setup()

	code := m.Run()
	if _, err := os.Stat(dbFile); err == nil {
		os.Remove(dbFile)
	}

	os.Exit(code)
}

func TestSignup(t *testing.T) {
	user := sqlDb.User{
		Username: "test1",
		Password: "pass1",
		Session:  "session",
		Email:    "test1@test.com",
	}

	assert.Nil(t, sqlDb.CreateUser(db, &user), "failed creating user")

	u, err := sqlDb.GetUser(db, user.Username)
	assert.Nil(t, err)
	assert.NotNil(t, u, "username is nil")

	user.Id = u.Id
	assert.EqualValues(t, &user, u)
}

func TestLogin(t *testing.T) {
	user := sqlDb.User{
		Username: "test",
		Password: "pass",
		Session:  "session",
		Email:    "test@test.com",
	}

	// login
	u, err := sqlDb.GetUser(db, user.Username)
	assert.ErrorIs(t, err, nil)
	assert.Nil(t, u)

	// signup
	err = sqlDb.CreateUser(db, &user)
	assert.Nil(t, err, nil)

	// login again
	u, err = sqlDb.GetUser(db, user.Username)
	assert.ErrorIs(t, err, nil)
	user.Id = u.Id
	assert.EqualValues(t, &user, u)
}
