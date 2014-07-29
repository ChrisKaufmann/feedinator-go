package main

import (
	"database/sql"
	"html"
	"html/template"
)

var (
	stmtAddEntry        *sql.Stmt
	stmtUpdateMarkEntry *sql.Stmt
	stmtUpdateReadEntry *sql.Stmt
	stmtSaveEntry       *sql.Stmt
	stmtGetEntryCount   *sql.Stmt
	entrySelectString   string
)

func init() {
	entrySelectString = " id,IFNULL(title,''),IFNULL(link,''),IFNULL(updated,''),marked,unread,feed_id,content,guid "
	stmtAddEntry = sth(db, "insert into ttrss_entries (updated,title,link,feed_id,marked,content,content_hash,unread,guid,user_name) values (NOW(),?,?,?,?,?,?,1,?,?)")
	stmtUpdateMarkEntry = sth(db, "update ttrss_entries set marked=? where id=?")
	stmtUpdateReadEntry = sth(db, "update ttrss_entries set unread=? where id=?")
	stmtSaveEntry = sth(db, "update ttrss_entries set title=?,link=?,updated=?,feed_id=?,marked=?,unread=? where id=? limit 1")
	stmtGetEntryCount = sth(db, "select count(id) from ttrss_entries")
}

type Entry struct {
	ID          int
	Evenodd     string
	Title       string
	Link        string
	Date        string
	FeedName    string
	Marked      string
	MarkSet     string
	FeedID      int
	Content     template.HTML
	ContentHash string
	Unread      bool
	ReadUnread  string
	GUID        string
}

func (e Entry) Normalize() Entry {
	e.Link = unescape(e.Link)
	e.Link = html.UnescapeString(e.Link)
	e.Title = unescape(e.Title)
	e.Title = html.UnescapeString(e.Title)
	e.FeedName = html.UnescapeString(e.FeedName)
	//sometimes there are duplicate encodings, replace &amp;#<something> with &#<something>
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
func getEntriesFromSql(s string) []Entry {
	var el []Entry
	var stmt = sth(db, s)
	rows, err := stmt.Query()
	if err != nil {
		err.Error()
		return el
	}
	var count int
	for rows.Next() {
		var e Entry
		var c string
		rows.Scan(&e.ID, &e.Title, &e.Link, &e.Date, &e.Marked, &e.Unread, &e.FeedID, &c, &e.GUID)
		e.Evenodd = evenodd(count)
		c = unescape(c)
		e.Content = template.HTML(html.UnescapeString(c))
		e.Link = html.UnescapeString(e.Link)
		e.Title = html.UnescapeString(e.Title)
		e.FeedName = e.Feed().Title
		e = e.Normalize()
		el = append(el, e)
		count = count + 1
	}
	return el
}
func allMarkedEntries() []Entry {
	sql := "select " + entrySelectString + " from ttrss_entries as e where e.user_name='" + userName + "' and e.marked=1"
	el := getEntriesFromSql(sql)
	return el
}
func (e Entry) Print() {
	print("ID:\t" + tostr(e.ID) + "\nTitle:\t" + e.Title + "\nLink:\t" + e.Link + "\nDate\t" + e.Date + "\nFeed_id:\t" + tostr(e.FeedID) + "\nMarked:\t" + e.Marked + "\nUnread:\t" + tostr(e.Unread) + "\nGuid:\t" + e.GUID + "\n")
}
func (e Entry) ViewMode() string {
	return e.Feed().ViewMode
}
func (e Entry) AutoscrollPX() int {
	return e.Feed().AutoscrollPX
}
func getEntriesCount() (c string, err error) {
	err = stmtGetEntryCount.QueryRow().Scan(&c)
	return c, err
}
func (e Entry) Feed() (f Feed) {
	f = getFeed(tostr(e.FeedID))
	return f
}
func (e Entry) Save() {
	if e.ID > 0 {
		stmtSaveEntry.Exec(e.Title, e.Link, e.Date, e.FeedID, e.Marked, e.Unread, e.ID)
	} else {
		_, err := stmtAddEntry.Exec(e.Title, e.Link, e.FeedID, e.Marked, tostr(e.Content), tostr(e.ContentHash), e.GUID, userName)
		if err != nil {
			err.Error()
		}
	}
}

func getEntry(id string) (e Entry) {
	if id == "" {
		return e
	}
	sql := "select " + entrySelectString + "from ttrss_entries where id='" + id + "'"
	el := getEntriesFromSql(sql)
	if len(el) > 0 {
		e = el[0]
		f := e.Feed()
		if f.UserName == userName {
			return e
		}
	}
	var badentry Entry
	return badentry
}
func markEntry(id string, m string) string {
	if id == "" {
		return ""
	}
	var ret string
	switch m {
	case "read":
		stmtUpdateReadEntry.Exec("0", id)
		e := getEntry(id)
		f := e.Feed()
		mc.Decrement("Category"+tostr(f.CategoryID)+"_UnreadCount", 1)
		mc.Decrement("Feed"+tostr(f.ID)+"_UnreadCount", 1)
		mc.Delete("Category" + tostr(f.CategoryID) + "_unreadentries")
		mc.Delete("Feed" + tostr(f.ID) + "_unreadentries")
		mc.Delete("Category" + tostr(f.CategoryID) + "_readentries")
		mc.Delete("Feed" + tostr(f.ID) + "_readentries")
	case "unread":
		stmtUpdateReadEntry.Exec("1", id)
		e := getEntry(id)
		f := e.Feed()
		mc.Increment("Category"+tostr(f.CategoryID)+"_UnreadCount", 1)
		mc.Increment("Feed"+tostr(f.ID)+"_UnreadCount", 1)
		mc.Delete("Category" + tostr(f.CategoryID) + "_unreadentries")
		mc.Delete("Feed" + tostr(f.ID) + "_unreadentries")
		mc.Delete("Category" + tostr(f.CategoryID) + "_readentries")
		mc.Delete("Feed" + tostr(f.ID) + "_readentries")
	case "marked":
		e := getEntry(id)
		f := e.Feed()
		mc.Delete("Feed" + tostr(e.FeedID) + "_markedentries")
		mc.Delete("Category" + tostr(f.CategoryID) + "_markedentries")
		stmtUpdateMarkEntry.Exec("1", id)
	case "unmarked":
		stmtUpdateMarkEntry.Exec("0", id)
	case "togglemarked":
		e := getEntry(id)
		f := e.Feed()
		stmtUpdateMarkEntry.Exec(toint(e.Marked)^1, id)
		en := getEntry(id)
		ret = "<img src='static/mark_" + en.MarkSet + ".png' alt='Set mark' onclick='javascript:toggleMark(" + id + ");'>\n"
		mc.Delete("Feed" + tostr(e.FeedID) + "_markedentries")
		mc.Delete("Category" + tostr(f.CategoryID) + "_markedentries")
	}
	return ret
}
