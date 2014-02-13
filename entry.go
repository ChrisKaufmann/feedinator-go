package main

import (
	"html"
	"database/sql"
	"html/template"
	"strconv"
)

var (
	stmtGetEntry        *sql.Stmt
	stmtUpdateMarkEntry *sql.Stmt
	stmtUpdateReadEntry *sql.Stmt
	stmtSaveEntry       *sql.Stmt
	stmtGetMarked       *sql.Stmt
	stmtCatEntries      *sql.Stmt
	stmtFeedEntries     *sql.Stmt
	stmtCatUnreadEntries *sql.Stmt
	stmtGetEntryCount	*sql.Stmt
)

func (e Entry) Save() {
	stmtSaveEntry.Exec(e.Title, e.Link, e.Date, e.FeedID, e.Marked, e.Unread, e.ID)
}

func init() {
	stmtGetEntry = sth(db, "select id,title,link,updated,feed_id,marked,content,unread from ttrss_entries where id= ?")
	stmtGetMarked = sth(db, "select e.id,IFNULL(e.title,''),IFNULL(e.link,''),IFNULL(e.updated,''),e.marked,e.unread,IFNULL(f.title,'') from ttrss_entries as e,ttrss_feeds as f where f.user_name= ? and e.marked=1")
	stmtUpdateMarkEntry = sth(db, "update ttrss_entries set marked=? where id=?")
	stmtUpdateReadEntry = sth(db, "update ttrss_entries set unread=? where id=?")
	stmtSaveEntry = sth(db, "update ttrss_entries set title=?,link=?,updated=?,feed_id=?,marked=?,unread=? where id=? limit 1")
	stmtFeedEntries = sth(db, "select e.id,IFNULL(e.title,''),IFNULL(e.link,''),IFNULL(e.updated,''),e.marked,e.unread,IFNULL(f.title,'') from ttrss_entries as e, ttrss_feeds as f where f.id = e.feed_id and f.id= ? and unread= ? and marked = ?")
	stmtCatEntries = sth(db, "select e.id,IFNULL(e.title,''),IFNULL(e.link,''),IFNULL(e.updated,''),e.marked,e.unread,IFNULL(f.title,'') from ttrss_entries e, ttrss_feeds as f where e.feed_id=f.id and f.category_id = ? and unread= ? and marked = ?")
	stmtCatUnreadEntries = sth(db, "select e.id,IFNULL(e.title,''),IFNULL(e.link,''),IFNULL(e.updated,''),e.marked,e.unread,IFNULL(f.title,'') from ttrss_entries e, ttrss_feeds as f where e.feed_id=f.id and f.category_id = ? and e.unread=1")
	stmtGetEntryCount = sth(db, "select count(id) from ttrss_entries")
}

type Entry struct {
	ID         int
	Evenodd    string
	Title      string
	Link       string
	Date       string
	FeedName   string
	Marked     string
	MarkSet    string
	FeedID     int
	Content    template.HTML
	Unread     bool
	ReadUnread string
}

func getCategoryUnread(id string) []Entry {
	rows, err := stmtCatUnreadEntries.Query(id)
	var el []Entry
	if err != nil {
		err.Error()
	}
	var count int
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.Title, &e.Link, &e.Date, &e.Marked, &e.Unread, &e.FeedName)
		e.Evenodd = evenodd(count)
		e=e.Normalize()
		el = append(el, e)
		count = count + 1
	}
	return el
}
func (e Entry) Normalize() Entry{
	e.Link = html.UnescapeString(e.Link)
	e.Title = html.UnescapeString(e.Title)
	e.FeedName = html.UnescapeString(e.FeedName)
	if e.Marked == "1" {
		e.MarkSet = "set"
	} else {
		e.MarkSet = "unset"
	}
	if e.Unread == true {
		e.ReadUnread = "unread"
	} else {
		e.ReadUnread = ""
	}
	return e
}
func entriesFromSql(s *sql.Stmt, id string, ur int, mkd string) []Entry {
	rows, err := s.Query(id, strconv.Itoa(ur), mkd)
	var el []Entry
	if err != nil {
		err.Error()
	}
	var count int
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.Title, &e.Link, &e.Date, &e.Marked, &e.Unread, &e.FeedName)
		e.Evenodd = evenodd(count)
		e=e.Normalize()
		el = append(el, e)
		count = count + 1
	}
	return el
}
func allMarkedEntries() []Entry {
	rows, err := stmtGetMarked.Query(userName)
	var el []Entry
	if err != nil {
		err.Error()
	}
	var count int
	for rows.Next() {
		var id string
		rows.Scan(&id)
		e := getEntry(id)
		e.Evenodd = evenodd(count)
		el = append(el, e)
		count = count + 1
	}
	return el
}
func (e Entry) ViewMode() string {
	f := getFeed(strconv.Itoa(e.FeedID))
	return f.ViewMode
}
func (e Entry) AutoscrollPX() int {
	f := getFeed(strconv.Itoa(e.FeedID))
	return f.AutoscrollPX
}
func getEntriesCount() (c string,err error) {
	err = stmtGetEntryCount.QueryRow().Scan(&c)
	return c,err
}
func getEntry(id string) Entry {
	//id,title,link,updated,feed_id,marked,content,unread
	var e Entry
	var c string
	err := stmtGetEntry.QueryRow(id).Scan(&e.ID, &e.Title, &e.Link, &e.Date, &e.FeedID, &e.Marked, &c, &e.Unread)
	if err != nil {
		err.Error()
	}
	e.Content = template.HTML(html.UnescapeString(c))
	e.Link = html.UnescapeString(e.Link)
	e.Title = html.UnescapeString(e.Title)
	if e.Marked == "1" {
		e.MarkSet = "set"
	} else {
		e.MarkSet = "unset"
	}
	if e.Unread == true {
		e.ReadUnread = "unread"
	} else {
		e.ReadUnread = ""
	}
	return e
}
func markEntry(id string, m string) string {
	var ret string
	switch m {
	case "read":
		stmtUpdateReadEntry.Exec("0", id)
	case "unread":
		stmtUpdateReadEntry.Exec("1", id)
	case "marked":
		stmtUpdateMarkEntry.Exec("1", id)
	case "unmarked":
		stmtUpdateMarkEntry.Exec("0", id)
	case "togglemarked":
		e := getEntry(id)
		stmtUpdateMarkEntry.Exec(toint(e.Marked)^1, id)
		en := getEntry(id)
		ret = "<img src='static/mark_" + en.MarkSet + ".png' alt='Set mark' onclick='javascript:toggleMark(" + id + ");'>\n"
	}
	return ret
}
