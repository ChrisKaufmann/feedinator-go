package feed

import (
	u "github.com/ChrisKaufmann/goutils"
	"strings"
	"database/sql"
	"html"
	"html/template"
	"github.com/golang/glog"
)

var (
	stmtAddEntry        *sql.Stmt
	stmtUpdateMarkEntry *sql.Stmt
	stmtUpdateReadEntry *sql.Stmt
	stmtSaveEntry       *sql.Stmt
	stmtGetEntryCount   *sql.Stmt
	entrySelectString   string
)

func Entryinit() {
	var err error
	entrySelectString = " id,IFNULL(title,''),IFNULL(link,''),IFNULL(updated,''),marked,unread,feed_id,content,guid "
	stmtAddEntry,err = u.Sth(db, "insert into ttrss_entries (updated,title,link,feed_id,marked,content,content_hash,unread,guid,user_name) values (NOW(),?,?,?,?,?,?,1,?,?)")
	if err != nil {glog.Fatalf("stmt: %s", err)}
	stmtUpdateMarkEntry,err = u.Sth(db, "update ttrss_entries set marked=? where id=?")
	if err != nil {glog.Fatalf("stmtUpdateMarkEntry: %s", err)}
	stmtUpdateReadEntry,err = u.Sth(db, "update ttrss_entries set unread=? where id=?")
	if err != nil {glog.Fatalf("stmtUpdateReadEntry: %s", err)}
	stmtSaveEntry,err = u.Sth(db, "update ttrss_entries set title=?,link=?,updated=?,feed_id=?,marked=?,unread=? where id=? limit 1")
	if err != nil {glog.Fatalf("stmtSaveEntry: %s", err)}
	stmtGetEntryCount,err = u.Sth(db, "select count(id) from ttrss_entries")
	if err != nil {glog.Fatalf("stmtGetEntryCount: %s", err)}
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
	var stmt,err = u.Sth(db, s)
	if err != nil {
		glog.Errorf("Error preparing statment '%s': %s", s, err)
	}
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
func AllMarkedEntries(userName string) []Entry {
	sql := "select " + entrySelectString + " from ttrss_entries as e where e.user_name='" + userName + "' and e.marked=1"
	el := getEntriesFromSql(sql)
	return el
}
func (e Entry) Print() {
	print("ID:\t" + u.Tostr(e.ID) + "\nTitle:\t" + e.Title + "\nLink:\t" + e.Link + "\nDate\t" + e.Date + "\nFeed_id:\t" + u.Tostr(e.FeedID) + "\nMarked:\t" + e.Marked + "\nUnread:\t" + u.Tostr(e.Unread) + "\nGuid:\t" + e.GUID + "\n")
}
func (e Entry) ViewMode() string {
	return e.Feed().ViewMode
}
func (e Entry) AutoscrollPX() int {
	return e.Feed().AutoscrollPX
}
func GetEntriesCount() (c string, err error) {
	err = stmtGetEntryCount.QueryRow().Scan(&c)
	return c, err
}
func (e Entry) Feed() (f Feed) {
	f = GetFeed(e.FeedID)
	return f
}
func (e Entry) Save(userName string) {
	if e.ID > 0 {
		stmtSaveEntry.Exec(e.Title, e.Link, e.Date, e.FeedID, e.Marked, e.Unread, e.ID)
	} else {
		_, err := stmtAddEntry.Exec(e.Title, e.Link, e.FeedID, e.Marked, u.Tostr(e.Content), u.Tostr(e.ContentHash), e.GUID, userName)
		if err != nil {
			err.Error()
		}
	}
}
func (e Entry) MarkRead() (err error) {
	if _, err = stmtUpdateReadEntry.Exec("0", e.ID); err != nil {
		glog.Errorf("stmtUpdateReadEntry.Exec(0,%v): %s", e.ID, err)
	}
	e.Feed().DecrementUnread()
	return err
}
func (e Entry) MarkUnread() (err error) {
	if _, err = stmtUpdateReadEntry.Exec("1", e.ID); err != nil {
		glog.Errorf("stmtUpdateReadEntry.Exec(1,%v): %s", e.ID, err)
	}
	e.Feed().IncrementUnread()
	return err
}
func (e Entry) Mark() (err error) {
	if _, err = stmtUpdateMarkEntry.Exec("1", e.ID); err != nil {
		glog.Errorf("stmtUpdateMarkEntry.Exec(1,%v): %s", e.ID, err)
		return err
	}
	e.Feed().ClearMarked()
	e.MarkSet="set"
	e.Marked="1"
	return err
}
func (e Entry) UnMark() (err error) {
	if _, err = stmtUpdateMarkEntry.Exec("0", e.ID); err != nil {
		glog.Errorf("stmtUpdateMarkEntry.Exec(0,%v): %s", e.ID, err)
		return err
	}
	e.Feed().ClearMarked()
	e.MarkSet="unset"
	e.Marked="0"
	return err
}
func (e Entry) ToggleMark() (retstr string, err error) {
	if e.Marked == "1" {
		return "unset",e.UnMark()
	}
	return "set",e.Mark()
}

func GetEntry(id string,userName string) (e Entry) {
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
		} else {
			glog.Errorf("f.Username(%s) does not match passed username(%s)", f.UserName, userName)
		}
	}
	var badentry Entry
	return badentry
}
/*
func MarkEntry(id string, m string, userName string) string {
	if id == "" {
		return ""
	}
	print("MarkEntry( %v, %s %s)", id, m, userName)
	var ret string
	switch m {
	case "togglemarked":
		e := GetEntry(id,userName)
		f := e.Feed()
		stmtUpdateMarkEntry.Exec(u.Toint(e.Marked)^1, id)
		en := GetEntry(id,userName)
		ret = "<img src='static/mark_" + en.MarkSet + ".png' alt='Set mark' onclick='javascript:toggleMark(" + id + ");'>\n"
		mc.Delete("Feed" + u.Tostr(e.FeedID) + "_markedentries")
		mc.Delete("Category" + u.Tostr(f.CategoryID) + "_markedentries")
	}
	return ret
}
*/
func unescape(s string) string {
    var codes = map[string]string{
        "&amp;":               "&",
        "&nbsp;":              " ",
        "&acirc;&#128;&#153;": "'",
    }
    for k, v := range codes {
        s = strings.Replace(s, k, v, -1)
    }
    return s
}
func evenodd(i int) string {
    if i%2 == 0 {
        return "even"
    }
    return "odd"
}
