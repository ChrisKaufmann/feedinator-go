package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Category struct {
	Name        string
	Description string
	UserName    string
	ID          int
	Unread      int
	Evenodd     string
	Class       string
}

var (
	stmtCatList               *sql.Stmt
	stmtCatUnread             *sql.Stmt
	stmtCatEntries            *sql.Stmt
	stmtGetCat                *sql.Stmt
	stmtGetCats               *sql.Stmt
	stmtNextCategoryEntry     *sql.Stmt
	stmtPreviousCategoryEntry *sql.Stmt
	stmtGetCatFeeds           *sql.Stmt
	stmtGetFeedsInCat         *sql.Stmt
	stmtSaveCat				  *sql.Stmt
)

func init() {
	stmtCatList = sth(db, "select name,id from ttrss_categories where user_name=?")
	stmtCatUnread = sth(db, "select count(e.id) as unread from ttrss_entries as e,ttrss_feeds as f where f.category_id= ? and e.feed_id=f.id and e.unread='1' order by e.id ASC")
	stmtCatEntries = sth(db, "select e.id from ttrss_entries as e, ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and e.feed_id=f.id and c.id = ? and unread= ? and marked = ? order by e.id ASC")
	stmtGetCatFeeds = sth(db, "select f.id from ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and c.id= ?")
	stmtGetCat = sth(db, "select name,user_name,description,id from ttrss_categories where id = ?")
	stmtGetCats = sth(db, "select name,user_name,description,id from ttrss_categories where user_name= ?")
	stmtNextCategoryEntry = sth(db, "select e.id from ttrss_entries as e,ttrss_feeds as f  where f.category_id=? and e.feed_id=f.id and e.id > ? order by e.id ASC limit 1")
	stmtPreviousCategoryEntry = sth(db, "select e.id from ttrss_entries as e, ttrss_feeds as f where f.category_id=? and e.feed_id=f.id and e.id<? order by e.id DESC limit 1")
	stmtSaveCat=sth(db,"update ttrss_categories set name=?,description=? where id=? limit 1")

}
func (c Category) Save() {
	stmtSaveCat.Exec(c.Name, c.Description, c.ID)
}

func getCat(id string) Category {
	var cat Category
	err := stmtGetCat.QueryRow(id).Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
	if err != nil {
		err.Error()
	}
	return cat
}
func getCategoryFeeds(id string) []Feed {
	var allFeeds []Feed
	rows, err := stmtGetCatFeeds.Query(id)
	if err != nil {
		err.Error()
		return allFeeds
	}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		allFeeds = append(allFeeds, getFeed(id))
	}
	return allFeeds
}
func getCategories() []Category {
	var allCats []Category
	rows, err := stmtGetCats.Query(userName)
	if err != nil {
		err.Error()
		return allCats
	}
	for rows.Next() {
		var cat Category
		rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
		cat.Unread = unreadCategoryCount(cat.ID)
		if cat.Unread > 0 {
			cat.Class = "oddUnread"
		} else {
			cat.Class = "odd"
		}
		allCats = append(allCats, cat)
	}
	return allCats
}
func unreadCategoryCount(id int) int {
	var count int
	err := stmtCatUnread.QueryRow(id).Scan(&count)
	if err != nil {
		err.Error()
	}
	return count
}
