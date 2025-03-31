package db

import (
	"database/sql"
	"errors"
	"math"
	"math/rand"
)

type User struct {
	Id       int
	Username string
	Password string
	Email    string
	Session  string
}

func InitDB(file string) (*sql.DB, error) {
	_db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
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

	return _db, nil
}

func CreateUser(db *sql.DB, userInfo *User) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into user(id, username, password, email, session) values (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(rand.Intn(math.MaxInt32), userInfo.Username, userInfo.Password, userInfo.Email,
		userInfo.Session); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func GetUser(db *sql.DB, Username string) (*User, error) {
	stmt, err := db.Prepare("select * from user where username = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res := stmt.QueryRow(Username)

	var user User
	if err := res.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.Session); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //no db error, Username doesn't exist
		}
		return nil, err
	}
	return &user, nil
}

func EmailExists(db *sql.DB, email string) (bool, error) {
	stmt, err := db.Prepare("select id from user where email = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res := stmt.QueryRow(email)

	var id int
	if err := res.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil //no db error, email doesn't exist
		}
		return false, err
	}
	return true, nil
}

func SetSession(db *sql.DB, username, session string) error {
	stmt, err := db.Prepare("UPDATE user SET session = ? where username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(session, username)
	if err != nil {
		return err
	}

	if aff, err := res.RowsAffected(); err != nil {
		return err
	} else if aff == 0 {
		return errors.New("no rows afftectd") // is an error kind of
	}
	return nil
}

func GetUserBySession(db *sql.DB, session string) (*User, error) {
	stmt, err := db.Prepare("SELECT * from user where session = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res := stmt.QueryRow(session)
	var user User
	if err := res.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.Session); err != nil {
		return nil, err
	}
	return &user, nil
}
