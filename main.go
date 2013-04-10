package main

import (
	"code.google.com/p/goauth2/oauth"
	"crypto/sha512"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/msbranco/goconfig"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"aggregator"
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
type Feed struct {
	ID             int
	Title          string
	UserName       string
	Unread         int
	Evenodd        string
	Class          string
	Url            string
	LastUpdated    string
	Public         string
	Expirey        string
	CategoryID     int
	ViewMode       string
	AutoscrollPX   int
	Exclude        string
	ErrorString    string
	ViewModeSelect template.HTML
	CategorySelect template.HTML
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

var (
	anum := aggregator.getId()
	userName                string
	cachefile               = "cache.json"
	indexHtml               = template.Must(template.ParseFiles("templates/index-nologin.html"))
	mainHtml                = template.Must(template.ParseFiles("templates/main.html"))
	categoryHtml            = template.Must(template.ParseFiles("templates/category.html"))
	feedHtml                = template.Must(template.ParseFiles("templates/feed.html"))
	feedHtmlSpaced          = template.Must(template.ParseFiles("templates/feed_spaced.html"))
	listEntryHtml           = template.Must(template.ParseFiles("templates/listentry.html"))
	feedMenuHtml            = template.Must(template.ParseFiles("templates/feed_menu.html"))
	catMenuHtml             = template.Must(template.ParseFiles("templates/category_menu.html"))
	entryLinkHtml           = template.Must(template.ParseFiles("templates/entry_link.html"))
	entryHtml               = template.Must(template.ParseFiles("templates/entry.html"))
	cookieName              = "feedinator_auth"
	viewModes               = [...]string{"Default", "Link", "Extended", "Proxy"}
	db                      *sql.DB
	stmtCatList             *sql.Stmt
	stmtCookieIns           *sql.Stmt
	stmtCatUnread           *sql.Stmt
	stmtFeedUnread          *sql.Stmt
	stmtGetUserId           *sql.Stmt
	stmtGetFeedsWithoutCats *sql.Stmt
	stmtCatEntries          *sql.Stmt
	stmtGetCatFeeds         *sql.Stmt
	stmtFeedEntries         *sql.Stmt
	stmtMarkedEntries       *sql.Stmt
	stmtGetCat              *sql.Stmt
	stmtGetFeed             *sql.Stmt
	stmtGetFeedsInCat       *sql.Stmt
	stmtGetEntry            *sql.Stmt
	stmtGetCats             *sql.Stmt
	stmtGetFeeds            *sql.Stmt
	stmtUpdateMarkEntry     *sql.Stmt
	stmtUpdateReadEntry     *sql.Stmt
	stmtNextCategoryEntry	*sql.Stmt
	stmtPreviousCategoryEntry	*sql.Stmt
	stmtNextFeedEntry		*sql.Stmt
	stmtPreviousFeedEntry	*sql.Stmt
	db_name                 string
	db_host                 string
	db_user                 string
	db_pass                 string
)
var oauthCfg = &oauth.Config{
	AuthURL:     "https://accounts.google.com/o/oauth2/auth",
	TokenURL:    "https://accounts.google.com/o/oauth2/token",
	RedirectURL: "http://dev.feedinator.com/oauth2callback",
	Scope:       "https://www.googleapis.com/auth/userinfo.profile",
	TokenCache:  oauth.CacheFile(cachefile),
}

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo"
const port = "9000"

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
	oauthCfg.ClientId, err = c.GetString("Google", "ClientId")
	oauthCfg.ClientSecret, err = c.GetString("Google", "ClientSecret")
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
	stmtGetFeedsWithoutCats, err = db.Prepare("select title, id from ttrss_feeds where user_name=? and category_id is NULL")
	if err != nil {
		err.Error()
	}
	stmtGetFeedsInCat, err = db.Prepare("select title, id from ttrss_feeds where user_name=? and category_id is ?")
	if err != nil {
		err.Error()
	}
	stmtCatEntries, err = db.Prepare("select e.id,e.title,e.updated,e.link,e.marked,f.title from ttrss_entries as e, ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and e.feed_id=f.id and c.id = ? and unread= ?")
	if err != nil {
		err.Error()
	}
	stmtMarkedEntries, err = db.Prepare("select e.id,e.title,e.updated,e.link,e.marked,f.title from ttrss_entries as e, ttrss_feeds as f where f.id=e.feed_id and  f.user_name = ? and e.marked=1")
	if err != nil {
		err.Error()
	}
	stmtGetCatFeeds, err = db.Prepare("select f.id from ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and c.id= ?")
	if err != nil {
		err.Error()
	}
	stmtFeedEntries, err = db.Prepare("select e.id,e.title,e.updated,e.link,e.marked,f.title from ttrss_entries as e, ttrss_feeds as f where e.feed_id=f.id and f.id = ? and unread= ?")
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
}

func main() {
	http.HandleFunc("/main", handleMain)
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)
	http.HandleFunc("/categoryList/", handleCategoryList)
	http.HandleFunc("/feedList/", handleFeedList)
	http.HandleFunc("/entry/mark/", handleMarkEntry)
	http.HandleFunc("/entry/", handleEntry)
	http.HandleFunc("/entries/", handleEntries)
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/menu/", handleMenu)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/favicon.ico", http.StripPrefix("/favicon.ico", http.FileServer(http.Dir("./static/favicon.ico"))))
	print("Listening on 127.0.0.1:9000\n")
	http.ListenAndServe("127.0.0.1:9000", nil)
}
func handleMarkEntry(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var retstr string
	a := strings.Split(r.URL.Path[len("/entry/mark/"):], "/")
	id := a[0]     //id of the entry, mark, or feed
	tomark := a[1] //mark read, unread, starred(marked)
	b := strings.Split(id, ",")
	for i := range b {
		retstr = markEntry(b[i], tomark)
	}
	fmt.Fprintf(w, retstr)

}
func handleEntry(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	a := strings.Split(r.URL.Path[len("/entry/"):], "/")
	id := a[0]

	e := getEntry(id)
	if e.ViewMode == "link" {
		e.Link = unescape(e.Link)
		entryLinkHtml.Execute(w, e)
	} else {
		entryHtml.Execute(w, e)
	}
	markEntry(id, "read")
}
func handleMenu(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	a := strings.Split(r.URL.Path[len("/menu/"):], "/")
	feedOrCat := a[0]
	id := a[1]
	if feedOrCat == "category" {
		cat := getCat(id)
		catMenuHtml.Execute(w, cat)
	}
	if feedOrCat == "feed" {
		f := getFeed(id)
		// Create the ViewModeSelect
		var optionHtml string
		for i := range viewModes {
			m := viewModes[i]
			lbl := m
			if strings.ToLower(m) == strings.ToLower(f.ViewMode) {
				lbl = "*" + m
			}
			optionHtml = optionHtml + "<option value='" + strings.ToLower(m) + "'>" + lbl + "\n"
		}
		//This prints the dropdown category select
		var catHtml string
		allthecats := getCategories()
		for i := range allthecats {
			cat := allthecats[i]
			if cat.ID == f.CategoryID {
				cat.Name = "*" + cat.Name
			}
			catHtml = catHtml + "<option value='" + strconv.Itoa(cat.ID) + "'>" + cat.Name + "\n"
		}
		f.ViewModeSelect = template.HTML(optionHtml)
		f.CategorySelect = template.HTML(catHtml)
		feedMenuHtml.Execute(w, f)
	}
}

//print the list of all feeds
func handleFeedList(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	fmt.Fprintf(w, "<ul class='feedList' id='feedList'>\n")
	allthefeeds := getFeeds()
	for i := range allthefeeds {
		f := allthefeeds[i]
		feedHtml.Execute(w, f)
	}
	fmt.Fprintf(w, "</ul><td align='right'><form name='add_feed_form'><input type='text' name='add_feed_text'><input type='button' value='Add' onclick='add_feed(this.form)'></form></td>\n")
}

//print the list of categories (possibly with feeds in that cat), then the uncategorized feeds
func handleCategoryList(w http.ResponseWriter, r *http.Request) {
	a := strings.Split(r.URL.Path[len("categoryList/"):], "/")
	currentCat := "0"
	if len(a) > 1 {
		currentCat = a[1]
	}
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	fmt.Fprintf(w, "<ul class='feedList' id='feedList'>\n")
	allthecats := getCategories()
	for i := range allthecats {
		cat := allthecats[i]
		categoryHtml.Execute(w, cat)
		fmt.Fprintf(w, "<br>\n")
		if strconv.Itoa(cat.ID) == currentCat {
			catFeeds := getCategoryFeeds(currentCat)
			for j := range catFeeds {
				feedHtmlSpaced.Execute(w, catFeeds[j])
			}
		}
	}
	fmt.Fprintf(w, "<br>")
	rows, err := stmtGetFeedsWithoutCats.Query(userName)
	if err != nil {
		fmt.Println(w, err)
		err.Error()
		return
	}
	for rows.Next() {
		var t string
		var id string
		rows.Scan(&t, &id)
		feed := getFeed(id)
		feedHtml.Execute(w, feed)
	}

	//print the footer for the categories list
	fmt.Fprintf(w, "</ul>\n<td align='right'>\n<form name='add_feed_form'>\n<input type='text' name='add_feed_text'>\n<input type='button' value='Add' onclick='add_feed(this.form)'>\n	</form>\n</td>\n")
}

//print the list of entries for the selected category, feed, or marked
func handleEntries(w http.ResponseWriter, r *http.Request) {
	var err error
	// format is /entries/{feed|category}/<id>/{read|unread|next|previous}[/{feed_id|cat_id}]
	a := strings.Split(r.URL.Path[len("/entries/"):], "/")
	feedOrCat := a[0]
	id := a[1]
	var ur int
	switch a[2] {
		case "read":
			ur = 0
		case "unread":
			ur = 1
		case "next":
			var retval string
			if feedOrCat == "feed" {
				stmtNextFeedEntry.QueryRow(a[3],id).Scan(&retval)
			} else {
				stmtNextCategoryEntry.QueryRow(a[3],id).Scan(&retval)
			}
			fmt.Fprintf(w,retval)
			return
		case "previous":
			var retval string
			if feedOrCat == "feed" {
				stmtPreviousFeedEntry.QueryRow(a[3],id).Scan(&retval)
			} else {
				stmtPreviousCategoryEntry.QueryRow(a[3],id).Scan(&retval)
			}
			fmt.Fprintf(w,retval)
			return
	}
	//print header for list
	//fmt.Fprintf(w, "<form action='backend.php' method='POST' id='entries_form'>\n<input type='hidden' name='op' value='mark_list_read'>\n
	fmt.Fprintf(w,"<form id='entries_form'><table class='headlinesList' id='headlinesList' width='100%'>")
	// templates/listentry.html
	var rows *sql.Rows
	if feedOrCat == "feed" {
		rows, err = stmtFeedEntries.Query(id, ur)
		if err != nil {
			fmt.Println(w, err)
			err.Error()
			return
		}
	}
	if feedOrCat == "category" {
		rows, err = stmtCatEntries.Query(id, ur)
		if err != nil {
			fmt.Println(w, err)
			err.Error()
			return
		}
	}
	if feedOrCat == "marked" {
		rows, err = stmtMarkedEntries.Query(userName)
		if err != nil {
			fmt.Println(w, err)
			err.Error()
			return
		}
	}
	var count int
	for rows.Next() {
		var entry Entry
		rows.Scan(&entry.ID, &entry.Title, &entry.Date, &entry.Link, &entry.Marked, &entry.FeedName)
		if entry.Marked == "1" {
			entry.MarkSet = "set"
		} else {
			entry.MarkSet = "unset"
		}
		entry.Evenodd = evenodd(count)
		entry.Title = unescape(entry.Title)
		listEntryHtml.Execute(w, entry)
		count = count + 1
	}

	//print footer for entries list
	fmt.Fprintf(w, "</form>\n</table>\n")
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	print(userName)
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := mainHtml.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		if err := indexHtml.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Redirect(w, r, "/main", http.StatusFound)
	}
}

// Start the authorization process
func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	print("In handleauth\n")
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

// Function that handles the callback from the Google server
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{oauth.Config: oauthCfg}

	// Exchange the received code for a token
	tok, _ := t.Exchange(code)
	{
		tokenCache := oauth.CacheFile(cachefile)

		err := tokenCache.PutToken(tok)
		if err != nil {
			log.Fatal("Cache write:", err)
		}
		log.Printf("Token is cached in %v\n", tokenCache)
	}

	// Skip TLS Verify
	t.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Make the request.
	req, err := t.Client().Get(profileInfoURL)
	if err != nil {
		log.Fatal("Request Error:", err)
	}
	defer req.Body.Close()
	body, _ := ioutil.ReadAll(req.Body)
	log.Println(string(body))
	//body.id is the google id to use
	//set a cookie with the id, and random hash. then save the id/hash pair to db for lookup
	var f interface{}
	err = json.Unmarshal(body, &f)
	m := f.(map[string]interface{})
	print(m["id"].(string))
	if err != nil {
		panic(err.Error())
	}
	var authString = randomString(64)
	_, err = stmtCookieIns.Exec(m["id"], hash(authString))

	if err != nil {
		panic(err.Error())
	}
	//set the cookie
	cookie := http.Cookie{Name: cookieName, Value: authString}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/main", http.StatusFound)
}
func loggedIn(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		//just means that the cookie doesn't exist or we couldn't read it
		return false
	}
	tok := cookie.Value
	tokHash := hash(tok)
	stmtGet, err := db.Prepare("select name from ttrss_sessions where userid = ?")
	if err != nil {
		panic(err.Error())
	}
	defer stmtGet.Close()
	var userId string
	err = stmtGet.QueryRow(tokHash).Scan(&userId) // WHERE number = 13
	if err != nil {
		return false //probably no rows in result set
	}

	if userId != "" {
		userName = userId
		return true
	} else {
		return false
	}
	return false
}
func randomString(l int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
func hash(s string) string {
	h := sha512.New()
	h.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
func evenodd(i int) string {
	if i%2 == 0 {
		return "even"
	}
	return "odd"
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
func getFeed(id string) Feed {
	var feed Feed
	err := stmtGetFeed.QueryRow(id).Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.Expirey, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString)
	if err != nil {
		err.Error()
	}
	feed.Unread = unreadFeedCount(feed.ID)
	if feed.Unread > 0 {
		feed.Class = "oddUnread"
	} else {
		feed.Class = "odd"
	}
	if feed.Title == "" {
		feed.Title = "--untitled--"
	}
	return feed
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
func getFeeds() []Feed {
	var allFeeds []Feed
	rows, err := stmtGetFeeds.Query(userName)
	if err != nil {
		err.Error()
		return allFeeds
	}
	for rows.Next() {
		var feed Feed
		rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.Expirey, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString)
		feed.Unread = unreadFeedCount(feed.ID)
		if feed.Unread > 0 {
			feed.Class = "oddUnread"
		} else {
			feed.Class = "odd"
		}
		allFeeds = append(allFeeds, feed)
	}
	return allFeeds
}
func unreadCategoryCount(id int) int {
	var count int
	err := stmtCatUnread.QueryRow(id).Scan(&count)
	if err != nil {
		err.Error()
	}
	return count
}
func unreadFeedCount(id int) int {
	var count int
	err := stmtFeedUnread.QueryRow(id).Scan(&count)
	if err != nil {
		err.Error()
	}
	return count
}
func toint(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
func unescape(s string) string {
	s = strings.Replace(s, "&#34;", "\"", -1)
	s = strings.Replace(s, "&#47;", "/", -1)
	return s
}
