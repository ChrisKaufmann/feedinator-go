package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/msbranco/goconfig"
)

var (
	db                        *sql.DB
	stmtGet                   *sql.Stmt
	stmtCatList               *sql.Stmt
	stmtCookieIns             *sql.Stmt
	stmtCatUnread             *sql.Stmt
	stmtFeedUnread            *sql.Stmt
	stmtGetUserId             *sql.Stmt
	stmtGetFeedsWithoutCats   *sql.Stmt
	stmtCatEntries            *sql.Stmt
	stmtGetCatFeeds           *sql.Stmt
	stmtFeedEntries           *sql.Stmt
	stmtMarkedEntries         *sql.Stmt
	stmtGetCat                *sql.Stmt
	stmtGetFeed               *sql.Stmt
	stmtGetFeedsInCat         *sql.Stmt
	stmtGetEntry              *sql.Stmt
	stmtGetCats               *sql.Stmt
	stmtGetFeeds              *sql.Stmt
	stmtUpdateMarkEntry       *sql.Stmt
	stmtUpdateReadEntry       *sql.Stmt
	stmtNextCategoryEntry     *sql.Stmt
	stmtPreviousCategoryEntry *sql.Stmt
	stmtNextFeedEntry         *sql.Stmt
	stmtPreviousFeedEntry     *sql.Stmt
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
	stmtCatList, err = db.Prepare("select name,id from ttrss_categories where user_name=?")
	if err != nil {
		err.Error()
	}
	stmtCookieIns, err = db.Prepare("INSERT INTO ttrss_sessions (name,userid) VALUES( ?, ? )") // ? = placeholder
	if err != nil {
		err.Error()
	}
	stmtCatUnread, err = db.Prepare("select count(ttrss_entries.id) as unread from ttrss_entries ,ttrss_feeds  where ttrss_feeds.category_id= ? and ttrss_entries.feed_id=ttrss_feeds.id and ttrss_entries.unread='1'")
	if err != nil {
		err.Error()
	}
	stmtFeedUnread, err = db.Prepare("select count(ttrss_entries.id) as unread from ttrss_entries where ttrss_entries.feed_id=? and ttrss_entries.unread='1'")
	if err != nil {
		err.Error()
	}
	stmtGetUserId, err = db.Prepare("select name from ttrss_sessions where userid = ?")
	if err != nil {
		err.Error()
	}
	stmtGetFeedsWithoutCats, err = db.Prepare("select id from ttrss_feeds where user_name=? and category_id is NULL")
	if err != nil {
		err.Error()
	}
	stmtGetFeedsInCat, err = db.Prepare("select title, id from ttrss_feeds where user_name=? and category_id is ?")
	if err != nil {
		err.Error()
	}
	stmtCatEntries, err = db.Prepare("select e.id from ttrss_entries as e, ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and e.feed_id=f.id and c.id = ? and unread= ?")
	if err != nil {
		err.Error()
	}
	stmtMarkedEntries, err = db.Prepare("select e.id from ttrss_entries as e, ttrss_feeds as f where f.id=e.feed_id and  f.user_name = ? and e.marked=1")
	if err != nil {
		err.Error()
	}
	stmtGetCatFeeds, err = db.Prepare("select f.id from ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and c.id= ?")
	if err != nil {
		err.Error()
	}
	stmtFeedEntries, err = db.Prepare("select e.id from ttrss_entries as e, ttrss_feeds as f where e.feed_id=f.id and f.id = ? and unread= ?")
	if err != nil {
		err.Error()
	}
	stmtGetCat, err = db.Prepare("select name,user_name,description,id from ttrss_categories where id = ?")
	if err != nil {
		err.Error()
	}
	stmtGetFeed, err = db.Prepare("select id,title,feed_url,last_updated,user_name,public,expirey,category_id,view_mode,autoscroll_px,exclude,error_string from ttrss_feeds where id = ?")
	if err != nil {
		err.Error()
	}
	stmtGetEntry, err = db.Prepare("select id,title,link,updated,feed_id,marked,content,unread from ttrss_entries where id= ?")
	if err != nil {
		err.Error()
	}
	stmtGetCats, err = db.Prepare("select name,user_name,description,id from ttrss_categories where user_name= ?")
	if err != nil {
		err.Error()
	}
	stmtGetFeeds, err = db.Prepare("select id, title, feed_url, last_updated, user_name,public, expirey, category_id, view_mode, autoscroll_px, exclude, error_string from ttrss_feeds where user_name = ?")
	if err != nil {
		err.Error()
	}
	stmtUpdateMarkEntry, err = db.Prepare("update ttrss_entries set marked=? where id=?")
	if err != nil {
		err.Error()
	}
	stmtUpdateReadEntry, err = db.Prepare("update ttrss_entries set unread=? where id=?")
	if err != nil {
		err.Error()
	}
	stmtNextCategoryEntry, err = db.Prepare("select e.id from ttrss_entries as e,ttrss_feeds as f  where f.category_id=? and e.feed_id=f.id and e.id > ? order by e.id ASC limit 1")
	if err != nil {
		err.Error()
	}
	stmtPreviousCategoryEntry, err = db.Prepare("select e.id from ttrss_entries as e, ttrss_feeds as f where f.category_id=? and e.feed_id=f.id and e.id<? order by e.id DESC limit 1")
	if err != nil {
		err.Error()
	}
	stmtNextFeedEntry, err = db.Prepare("select id from ttrss_entries where feed_id=? and id > ? limit 1")
	if err != nil {
		err.Error()
	}
	stmtPreviousFeedEntry, err = db.Prepare("select id from ttrss_entries where feed_id=? and id<? order by id DESC limit 1")
	if err != nil {
		err.Error()
	}
	stmtGet, err = db.Prepare("select name from ttrss_sessions where userid = ?")
	if err != nil {
		err.Error()
	}
}
