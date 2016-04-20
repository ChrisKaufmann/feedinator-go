package auth

//  auth/user.go

import (
	"../feed"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang/glog"
)

type User struct {
	ID    string
	Email string
}

//object functions
func (us User) String() string {
	return fmt.Sprintf("ID: %s, Email: %s", us.ID, us.Email)
}
func (us User) AddSession(sh string) (err error) {
	_, err = stmtCookieIns.Exec(us.ID, sh)
	if err != nil {
		glog.Errorf("user.AddSession(%s)stmtCookieIns(%s,%s):%s", us, us.ID, sh, err)
	}
	return err
}
func (us User) Categories() []feed.Category {
	return feed.GetCategories(us.Email)
}
func (us User) Feeds() (fl []feed.Feed) {
	return feed.GetFeeds(us.Email)
}

//Non object functions
func UserExists(email string) (exists bool) {
	var uid string
	err := stmtGetUserID.QueryRow(email).Scan(&uid)
	switch {
	case err == sql.ErrNoRows:
		exists = false
	case err != nil:
		glog.Errorf("UserExists():stmtGetUserID(%s): %s", email, err)
		exists = false
	default:
		exists = true
	}
	return exists
}
func AddUser(e string) (us User, err error) {
	if UserExists(e) {
		err = stmtGetUserID.QueryRow(e).Scan(&us.ID)
		return GetUserByEmail(e)
	}
	_, err = stmtInsertUser.Exec(e, e)
	if err != nil {
		glog.Errorf("AddUser(%s): %s", e, err)
		return us, err
	}
	us.ID = e
	us.Email = e
	return us, err
}
func GetUserByEmail(e string) (us User, err error) {
	if !UserExists(e) {
		err = errors.New("User Doesn't exist")
		return us, err
	}
	err = stmtGetUserID.QueryRow(e).Scan(&us.ID)
	if err != nil {
		glog.Errorf("GetUserByEmail()stmtGetUserID(%s): %s", e, err)
	}
	us.Email = e
	return us, err
}
func GetUserBySession(s string) (us User, err error) {
	err = stmtGetUserBySession.QueryRow(s).Scan(&us.ID, &us.Email)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("No valid session")
		return us, err
	case err != nil:
		glog.Errorf("GetUserBySession():stmtGetUserBySession(%s): %s", s, err)
		return us, err
	}
	return us, err
}
func SessionExists(s string) (e bool) {
	//	stmtSessionExists, err := u.Sth(db, "select user_id from sessions where session_hash=?"

	var uid string
	err := stmtSessionExists.QueryRow(s).Scan(&uid)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		glog.Errorf("SessionExists():stmtSessionExists(%s): %s", s, err)
		return false
	default:
		return true
	}
	return e
}
func GetAllUsers() (ul []User, err error) {
	rows, err := stmtGetAllUsers.Query()
	if err != nil {
		glog.Errorf("stmtGetAllUsers.Query(): %s", err)
		return ul, err
	}
	for rows.Next() {
		var us User
		rows.Scan(&us.ID, us.Email)
		ul = append(ul, us)
	}
	return ul, err
}
