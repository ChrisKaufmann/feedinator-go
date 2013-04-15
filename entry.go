package main

import (
	"database/sql"
	"html/template"
	"strconv"
)

var (
	stmtGetEntry        *sql.Stmt
	stmtUpdateMarkEntry *sql.Stmt
	stmtUpdateReadEntry *sql.Stmt
	stmtSaveEntry		*sql.Stmt
	stmtGetMarked		*sql.Stmt
)

func (e Entry)Save () {
	stmtSaveEntry.Exec(e.Title,e.Link,e.Date,e.FeedID,e.Marked,e.Unread,e.ID)
}

func init() {
	stmtGetEntry=sth(db,"select id,title,link,updated,feed_id,marked,content,unread from ttrss_entries where id= ?")
	stmtGetMarked=sth(db,"select e.id,e.title,e.link,e.updated,e.feed_id,e.marked,e.content,e.unread from ttrss_entries as e,ttrss_feeds as f where f.user_name= ? and e.marked=1")
	stmtUpdateMarkEntry=sth(db,"update ttrss_entries set marked=? where id=?")
	stmtUpdateReadEntry=sth(db,"update ttrss_entries set unread=? where id=?")
	stmtSaveEntry=sth(db,"update ttrss_entries set title=?,link=?,updated=?,feed_id=?,marked=?,unread=? where id=? limit 1")
}

type Entry struct {
	ID       int
	Evenodd  string
	Title    string
	Link     string
	Date     string
	FeedName string
	ViewMode string
	Marked   string
	MarkSet  string
	FeedID   int
	Content  template.HTML
	Unread   bool
}

func entriesFromSql(s *sql.Stmt, id string, ur int,mkd int) []Entry {
	rows, err := s.Query(id, strconv.Itoa(ur), mkd)
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
func allMarkedEntries() []Entry {
	rows,err := stmtGetMarked.Query(userName)
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
func getEntry(id string) Entry {
	//id,title,link,updated,feed_id,marked,content,unread
	var e Entry
	var c string
	err := stmtGetEntry.QueryRow(id).Scan(&e.ID, &e.Title, &e.Link, &e.Date, &e.FeedID, &e.Marked, &c, &e.Unread)
	if err != nil {
		err.Error()
	}
	if e.Marked == "1" {
		e.MarkSet = "set"
	} else {
		e.MarkSet = "unset"
	}
	f := getFeed(strconv.Itoa(e.FeedID))
	e.Content = template.HTML(unescape(c))
	e.Link = unescape(e.Link)
	e.Title = unescape(e.Title)
	e.FeedName = f.Title
	e.ViewMode = f.ViewMode
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