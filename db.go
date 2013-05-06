package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/msbranco/goconfig"
)

var (
	db                        *sql.DB
	stmtGet                   *sql.Stmt
	stmtCookieIns             *sql.Stmt
	stmtGetUserId             *sql.Stmt
	stmtMarkedEntries         *sql.Stmt
	db_name                   string
	db_host                   string
	db_user                   string
	db_pass                   string
)

func init() {
	var err error
	c, err := goconfig.ReadConfigFile("config")
	if err != nil {
		err.Error()
	}
	db_name, err = c.GetString("DB", "db")
	if err != nil {
		err.Error()
	}
	db_host, err = c.GetString("DB", "host")
	if err != nil {
		err.Error()
	}
	db_user, err = c.GetString("DB", "user")
	if err != nil {
		err.Error()
	}
	db_pass, err = c.GetString("DB", "pass")
	if err != nil {
		err.Error()
	}
	db, err = sql.Open("mysql", db_user+":"+db_pass+"@"+db_host+"/"+db_name)
	if err != nil {
		panic(err)
	}
	stmtCookieIns=sth(db,"INSERT INTO ttrss_sessions (name,userid) VALUES( ?, ? )") // ? = placeholder
	stmtGetUserId=sth(db,"select name from ttrss_sessions where userid = ?")
	stmtGetFeedsInCat=sth(db,"select title, id from ttrss_feeds where user_name= ? and category_id = ?")
	stmtMarkedEntries=sth(db,"select e.id from ttrss_entries as e, ttrss_feeds as f where f.id=e.feed_id and  f.user_name = ? and e.marked = ?")
	stmtGet=sth(db,"select u.username from users as u, ttrss_sessions as s where s.userid = ? and s.name=u.email")
}
