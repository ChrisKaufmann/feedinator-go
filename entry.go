package main

import (
	"database/sql"
	"html/template"
	"strconv"
)

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

func entriesFromSql(s *sql.Stmt, id string, ur int) []Entry {
	rows, err := s.Query(id, ur)
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
