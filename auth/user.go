package auth
//  auth/user.go

import (
	"database/sql"
	"errors"
	"github.com/golang/glog"
	"fmt"
)

type User struct {
	ID		string
	Email	string
}

//object functions
func (us User) String() (string) {
    return fmt.Sprintf("ID: %s, Email: %s",us.ID,us.Email)
}
func (us User) AddSession(sh string) (err error) {
	_,err = stmtCookieIns.Exec(us.ID, sh)
	if err != nil {glog.Errorf("user.AddSession(%s)stmtCookieIns(%s,%s):%s",us,us.ID,sh,err) }
	return err
}

//Non object functions
func UserExists(email string)(exists bool) {
    var uid string
    err := stmtGetUserID.QueryRow(email).Scan(&uid)
    switch {
        case err == sql.ErrNoRows:
            exists = false
        case err != nil:
			glog.Errorf("UserExists():stmtGetUserID(%s): %s",email,err)
			exists=false
        default:
            exists = true
    }
    return exists
}
func AddUser(e string)(us User, err error) {
    if UserExists(e) {
        err = stmtGetUserID.QueryRow(e).Scan(&us.ID)
		return GetUserByEmail(e)
    }
    _, err = stmtInsertUser.Exec(e,e)
    if err != nil {
		glog.Errorf("AddUser(%s): %s", e, err)
		return us,err
	}
    us.ID=e
	us.Email=e
    return us, err
}
func GetUserByEmail(e string)(us User, err error) {
    if !UserExists(e) {
        err=errors.New("User Doesn't exist")
		glog.Errorf("GetUserByEmail(%s): %s",e,err)
        return us, err
    }
    err = stmtGetUserID.QueryRow(e).Scan(&us.ID)
    if err != nil {
		glog.Errorf("GetUserByEmail()stmtGetUserID(%s): %s",e,err)
	}
	us.Email=e
    return us, err
}
func GetUserBySession(s string)(us User, err error) {
//	stmtGetUserByHash, err = u.Sth(db, "select user.id, user.email from user, sessions where user.id=sessions.user_id and sessions.session_hash=?")
//	stmtGetUser,err = us.Sth(db, "select user_id from sessions as s where s.session_hash = ?")
	err = stmtGetUserBySession.QueryRow(s).Scan(&us.ID, &us.Email)
	switch {
		case err == sql.ErrNoRows:
			err = errors.New("No valid session")
			return us, err
		case err != nil:
			glog.Errorf("GetUserBySession():stmtGetUserBySession(%s): %s",s,err)
			return us,err
	}
	return us,err
}
func GetUserByShared(s string)(us User, err error) {
	err = stmtGetUserByShared.QueryRow(s).Scan(&us.ID, &us.Email)
	switch {
		case err == sql.ErrNoRows:
			err = errors.New("No valid session")
			return us, err
		case err != nil:
			glog.Errorf("GetUserBySession():stmtGetUserBySession(%s): %s",s,err)
			return us,err
	}
	return us,err
}
func SessionExists(s string)(e bool) {
//	stmtSessionExists, err := u.Sth(db, "select user_id from sessions where session_hash=?"

	var uid string
	err := stmtSessionExists.QueryRow(s).Scan(&uid)
	switch {
		case err == sql.ErrNoRows:
			return false
		case err != nil:
			glog.Errorf("SessionExists():stmtSessionExists(%s): %s",s,err)
			return false
		default:
			return true
	}
	return e
}
