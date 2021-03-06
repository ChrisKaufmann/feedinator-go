package feed

import (
	"database/sql"
	"fmt"
	u "github.com/ChrisKaufmann/goutils"
	"github.com/golang/glog"
	"html"
	"html/template"
	"net/url"
	"strings"
)

var (
	stmtAddEntry        *sql.Stmt
	stmtUpdateMarkEntry *sql.Stmt
	stmtUpdateReadEntry *sql.Stmt
	stmtSaveEntry       *sql.Stmt
	stmtGetEntryCount   *sql.Stmt
	stmtGetEntry        *sql.Stmt
	entrySelectString   string
)

func Entryinit() {
	var err error
	entrySelectString = " id,IFNULL(title,''),IFNULL(link,''),IFNULL(updated,''),IFNULL(marked,0),IFNULL(unread,1),feed_id,content,guid "
	stmtAddEntry, err = u.Sth(db, "insert into ttrss_entries (updated,title,link,feed_id,marked,content,content_hash,unread,guid,user_name) values (NOW(),?,?,?,?,?,?,1,?,?)")
	if err != nil {
		glog.Fatalf("stmt: %s", err)
	}
	stmtUpdateMarkEntry, err = u.Sth(db, "update ttrss_entries set marked=? where id=?")
	if err != nil {
		glog.Fatalf("stmtUpdateMarkEntry: %s", err)
	}
	stmtUpdateReadEntry, err = u.Sth(db, "update ttrss_entries set unread=? where id=?")
	if err != nil {
		glog.Fatalf("stmtUpdateReadEntry: %s", err)
	}
	stmtSaveEntry, err = u.Sth(db, "update ttrss_entries set title=?,link=?,updated=?,feed_id=?,marked=?,unread=?,content=? where id=? limit 1")
	if err != nil {
		glog.Fatalf("stmtSaveEntry: %s", err)
	}
	stmtGetEntryCount, err = u.Sth(db, "select count(id) from ttrss_entries")
	if err != nil {
		glog.Fatalf("stmtGetEntryCount: %s", err)
	}
	stmtGetEntry, err = u.Sth(db, "select "+entrySelectString+"from ttrss_entries where id=?")
	if err != nil {
		glog.Fatalf("stmtGetEntry: %s", err)
	}
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
type EntryList []Entry

func (el EntryList) Len() int           { return len(el) }
func (el EntryList) Less(i, j int) bool { return el[i].ID < el[j].ID }
func (el EntryList) Swap(i, j int)      { el[i], el[j] = el[j], el[i] }

func GetEntry(id string, userName string) (e Entry) {
	if id == "" || userName == "" {
		glog.Errorf("No id(%s) or userName(%s) passed to GetEntry", id, userName)
		return e
	}
	el, err := getEntriesFromSthP(stmtGetEntry, id)
	if err != nil {
		glog.Errorf("getEntriesFromSthP(%s): %s", id, err)
		var e Entry
		return e
	}

	//	return el[0]

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
func (e Entry) String() string {
	return fmt.Sprintf("ID: %v\nTitle: %s\nLink: %s\nDate: %s\nMarked: %s\nMarkSet: %s\nFeedID: %v\nContent: %s\n Unread: %t\n", e.ID, e.Title, e.Link, e.Date, e.Marked, e.MarkSet, e.FeedID, e.Content, e.Unread)
}
func (e Entry) Print() {
	fmt.Printf("ID: %v\nFeed ID: %s\nTitle: %s\nLink: %s\nDate: %s\nMarked: %t\nUnread: %t\nGUID: %s\nContent: %s\n", e.ID, e.FeedID, e.Title, e.Link, e.Date, e.Marked, e.Unread, e.GUID, e.Content)
}
func (e Entry) ViewMode() string {
	return e.Feed().ViewMode
}
func (e Entry) AutoscrollPX() int {
	return e.Feed().AutoscrollPX
}
func (e Entry) Feed() (f Feed) {
	if e.FeedID < 1 {
		return f
	}
	f, err := GetFeed(e.FeedID)
	if err != nil {
		glog.Errorf("GetFeed(%v): %s", e.FeedID, err)
	}
	return f
}
func (e Entry) Save(userName string) (err error) {
	if e.ID > 0 {
		if e.Title == "" {
			e.Title = "&nbsp"
		}
		if e.Link == "" {
			e.Link = ""
		}
		if e.Date == "" {
			e.Date = ""
		}
		if e.FeedID == 0 {
			e.FeedID = 0
		}
		if e.Marked == "" {
			e.Marked = "0"
		}
		if u.Tostr(e.Content) == "" {
			e.Content = template.HTML("")
		}
		var unread string = "1"
		if e.Unread == false {
			unread = "0"
		}
		e.Link = strings.TrimSpace(e.Link)
		_, err = stmtSaveEntry.Exec(e.Title, e.Link, e.Date, e.FeedID, e.Marked, unread, u.Tostr(e.Content), e.ID)
		if err != nil {
			glog.Errorf("stmtSaveEntry.Exec(%s,%s,%s,%s,%s,%s,%s,%s): %s", e.Title, e.Link, e.Date, e.FeedID, e.Marked, unread, u.Tostr(e.Content), e.ID, err)
			return err
		}
	} else {
		_, err := stmtAddEntry.Exec(e.Title, e.Link, e.FeedID, e.Marked, u.Tostr(e.Content), u.Tostr(e.ContentHash), e.GUID, userName)
		if err != nil {
			glog.Errorf("stmtAddEntry: %s", err)
			return err
		}
	}
	return err
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
	e.MarkSet = "set"
	e.Marked = "1"
	return err
}
func (e Entry) UnMark() (err error) {
	if _, err = stmtUpdateMarkEntry.Exec("0", e.ID); err != nil {
		glog.Errorf("stmtUpdateMarkEntry.Exec(0,%v): %s", e.ID, err)
		return err
	}
	e.Feed().ClearMarked()
	e.MarkSet = "unset"
	e.Marked = "0"
	return err
}
func (e Entry) ToggleMark() (retstr string, err error) {
	if e.Marked == "1" {
		return "unset", e.UnMark()
	}
	return "set", e.Mark()
}
func (e Entry) ProxyLink() (h template.HTML, err error) {
	//Retrieve url content
        //	res, err := http.Get(html.UnescapeString(e.Link))
	//Get the actual domain and attempt to replace any img links that are relative
	e.Link = strings.TrimSpace(e.Link)
	url, err := url.Parse(e.Link)
	if err != nil {
		glog.Errorf("url.Parse(%s): %s", e.Link, err)
	}

	link := fmt.Sprintf("https://proxy.chriskaufmann.com/nph-proxy.pl/en/00/%s/%s%s", url.Scheme, url.Host, url.RequestURI() )
	fmt.Printf("link: %s", link)

	ht := fmt.Sprintf("<iframe id='view_iframe' src=\"%s\" style='overflow:auto;height:1000%%;frameborder:0;width:100%%' ></iframe>", link)
	return template.HTML(ht), err
/*
	res, err := http.Get(html.UnescapeString(link))
	if err != nil {
		glog.Errorf("http.Get(html.UnescapeString(%s)): %s", link, err)
		return h, err
	}
*/

/*
	client := &http.Client{}

	link := fmt.Sprintf("https://proxy.chriskaufmann.com/nph-proxy.pl/en/20/%s/%s%s",  url.Scheme, url.Host, url.RequestURI() )
	fmt.Printf("link: %s", link)
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		glog.Errorf("http.NewRuquest(%s): %s", e.Link, err)
	}
	req.SetBasicAuth(proxy_username, proxy_pass)
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("http.Get(%s): %s", e.Link, err)
		return h, err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("ioutil.ReadAll(): %s", err)
		return h, err
	}
	res.Body.Close()
	h = template.HTML(content)
	return h, err

*/
	/*
		ht := fmt.Sprintf("<iframe id='view_iframe' src=\"%s\" style='overflow:auto;height:1000%%;frameborder:0;width:100%%' ></iframe>", link)
		return template.HTML(ht), err

		client := &http.Client{}
		req, err := http.NewRequest("GET", e.Link, nil)
		if err != nil {
			glog.Errorf("http.NewRuquest(%s): %s", e.Link, err)
		}
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.89 Safari/537.36")
		res, err := client.Do(req)
		if err != nil {
			glog.Errorf("http.Get(%s): %s", e.Link, err)
			return h, err
		}
	*/

/*
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("ioutil.ReadAll(): %s", err)
		return h, err
	}
	res.Body.Close()
	c := fmt.Sprintf("url: %s\n Content: %s", e.Link, content)
	imgregex, err := regexp.Compile("img\\s+src\\s{0,}=\\s{0,}(['\"])/")
	if err != nil {
		glog.Errorf("regexp.compile(): %s")
	}
	fmt.Printf("host: %s", url.Host)
	c = imgregex.ReplaceAllString(c, "img src=$1")

	excludes := []string{"<head(.|\n)*?/head>", "<script(.|\n)*?/script>", "<noscript(.|\n)*?/noscript>"}
	for _, e := range excludes {
		reg, err := regexp.CompilePOSIX(e)
		if err != nil {
			glog.Errorf("rgexp.compile: %s", err)
			return h, err
		}
		c = reg.ReplaceAllString(c, " ")
	}
	fmt.Printf("html: %s", fmt.Sprintf(" %s ", c))
	h = template.HTML(c)
	return h, err
*/
}

func getEntriesFromSql(s string) []Entry {
	var el []Entry
	var stmt, err = u.Sth(db, s)
	if err != nil {
		glog.Errorf("Error preparing statment '%s': %s", s, err)
		return el
	}
	el, err = getEntriesFromSth(stmt)
	return el
}
func getEntriesFromSthP(stmt *sql.Stmt, p string) (el []Entry, err error) {
	rows, err := stmt.Query(p)
	if err != nil {
		glog.Errorf("getEntriesFromSthP.Query(%s): %s", p, err)
		return el, err
	}
	return getEntriesFromRows(rows)
}
func getEntriesFromSth(stmt *sql.Stmt) (el []Entry, err error) {
	rows, err := stmt.Query()
	if err != nil {
		glog.Errorf("stmt.Query() %s", err)
		return el, err
	}
	return getEntriesFromRows(rows)
}
func getEntriesFromRows(rows *sql.Rows) (el []Entry, err error) {
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
	return el, err
}

func AllMarkedEntries(userName string) []Entry {
	sql := "select " + entrySelectString + " from ttrss_entries as e where e.user_name='" + userName + "' and e.marked=1"
	el := getEntriesFromSql(sql)
	return el
}
func GetEntriesCount() (c string, err error) {
	err = stmtGetEntryCount.QueryRow().Scan(&c)
	return c, err
}

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
