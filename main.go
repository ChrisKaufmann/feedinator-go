package main

import (
	"github.com/msbranco/goconfig"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"net/url"
)

var (
	userName       string
	cachefile      = "cache.json"
	indexHtml      = template.Must(template.ParseFiles("templates/index-nologin.html"))
	mainHtml       = template.Must(template.ParseFiles("templates/main.html"))
	categoryHtml   = template.Must(template.ParseFiles("templates/category.html"))
	feedHtml       = template.Must(template.ParseFiles("templates/feed.html"))
	feedHtmlSpaced = template.Must(template.ParseFiles("templates/feed_spaced.html"))
	listEntryHtml  = template.Must(template.ParseFiles("templates/listentry.html"))
	feedMenuHtml   = template.Must(template.ParseFiles("templates/feed_menu.html"))
	catMenuHtml    = template.Must(template.ParseFiles("templates/category_menu.html"))
	entryLinkHtml  = template.Must(template.ParseFiles("templates/entry_link.html"))
	entryHtml      = template.Must(template.ParseFiles("templates/entry.html"))
	menuDropHtml   = template.Must(template.ParseFiles("templates/menu_dropdown.html"))
	cookieName     = "feedinator_auth"
	viewModes      = [...]string{"Default", "Link", "Extended", "Proxy"}
	port           string
)

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo"

func init() {
	var err error
	c, err := goconfig.ReadConfigFile("config")
	if err != nil {
		err.Error()
	}
	port, err = c.GetString("Web", "port")
	if err != nil {
		err.Error()
	}
}

func main() {
	http.HandleFunc("/main", handleMain)
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)
	http.HandleFunc("/categoryList/", handleCategoryList)
	http.HandleFunc("/category/", handleCategory)
	http.HandleFunc("/feed/list/", handleFeedList)
	http.HandleFunc("/feed/new/", handleNewFeed)
	http.HandleFunc("/feed/", handleFeed)
	http.HandleFunc("/entry/mark/", handleMarkEntry)
	http.HandleFunc("/entry/", handleEntry)
	http.HandleFunc("/entries/", handleEntries)
	http.HandleFunc("/menu/select/", handleSelectMenu)
	http.HandleFunc("/menu/", handleMenu)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/favicon.ico", http.StripPrefix("/favicon.ico", http.FileServer(http.Dir("./static/favicon.ico"))))
	http.HandleFunc("/", handleRoot)

	print("Listening on 127.0.0.1:"+port+"\n")
	http.ListenAndServe("127.0.0.1:"+port, nil)
}
func handleCategory(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	var todo string
	var val string
	pathVars(r, "/category/", &id, &todo, &val)
	switch todo {
	case "new":
		var c Category
		c.Name = val
		c.Insert()
		fmt.Fprintf(w, "Added")
	case "name":
		c := getCat(id)
		c.Name = val
		c.Save()
		fmt.Fprintf(w, id+"Renamed: "+val)
	case "desc":
		c := getCat(id)
		c.Description = val
		c.Save()
		fmt.Fprintf(w, "Desc: "+val)
	case "delete":
		c := getCat(id)
		c.Delete()
		fmt.Fprintf(w, "Deleted")
	}
}
func handleNewFeed(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	formurl := r.FormValue("url")
	var f Feed
	f.Url = formurl
	f.UserName = userName
	purl,_ := url.Parse(formurl)
	f.Title=purl.Host
	f.Insert()
	fmt.Fprintf(w, "Added")
}
func handleFeed(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	var todo string
	var val string
	pathVars(r, "/feed/", &id, &todo, &val)
	f := getFeed(id)
	if f.UserName != userName {
		fmt.Fprintf(w, "Auth err")
		return
	}
	switch todo {
	case "name":
		f.Title = val
		f.Save()
		fmt.Fprintf(w, "Name: "+val)
	case "link":
		url := r.FormValue("url")
		f.Url = url
		f.Save()
		fmt.Fprintf(w, f.Url)
	case "expirey":
		f.Expirey = val
		f.Save()
		fmt.Fprintf(w, "Expirey: "+val)
	case "autoscroll":
		f.AutoscrollPX = toint(val)
		f.Save()
		fmt.Fprintf(w, "Autoscroll: "+val)
	case "exclude":
		f.Exclude = val
		f.Save()
		fmt.Fprintf(w, "Exclude saved")
	case "category":
		f.CategoryID = toint(val)
		f.Save()
		c:=getCat(val)
		fmt.Fprintf(w, "Category: "+c.Name)
	case "view_mode":
		f.ViewMode = val
		f.Save()
		fmt.Fprintf(w, "View Mode: "+val)
	case "delete":
		f.Delete()
		fmt.Fprintf(w, "Deleted")
	}
	return
}
func handleMarkEntry(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var retstr string
	var id string
	var tomark string
	pathVars(r, "/entry/mark/", &id, &tomark)
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
	var id string
	pathVars(r, "/entry/", &id)

	e := getEntry(id)
	f := getFeed(tostr(e.FeedID))
	e.FeedName = f.Title
	if e.ViewMode() == "link" {
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
	var feedOrCat string
	var id string
	pathVars(r, "/menu/", &feedOrCat, &id)
	if feedOrCat == "category" {
		cat := getCat(id)
		catMenuHtml.Execute(w, cat)
	}
	if feedOrCat == "feed" {
		f := getFeed(id)
		setSelects(&f)
		feedMenuHtml.Execute(w, f)
	}
}
func handleSelectMenu(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	pathVars(r, "/menu/select/", &id)
	f := getFeed(id)
	setSelects(&f)
	menuDropHtml.Execute(w, f)
}
func setSelects(f *Feed) {
	var catHtml string
	var optionHtml string
	for i := range viewModes {
		m := viewModes[i]
		lbl := m
		if strings.ToLower(m) == strings.ToLower(f.ViewMode) {
			lbl = "*" + m
		}
		optionHtml = optionHtml + "<option value='" + strings.ToLower(m) + "'>" + lbl + "\n"
	}
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
}

//print the list of categories (possibly with feeds in that cat), then the uncategorized feeds
func handleCategoryList(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var currentCat string
	pathVars(r, "/categoryList/", &currentCat)
	fmt.Fprintf(w, "<ul class='feedList' id='feedList'>\n")
	allthecats := getCategories()
	for i := range allthecats {
		cat := allthecats[i]
		categoryHtml.Execute(w, cat)
		fmt.Fprintf(w, "<br>\n")
		//print the feeds under the currently selected category
		if strconv.Itoa(cat.ID) == currentCat {
			catFeeds := getCategoryFeeds(currentCat)
			for j := range catFeeds {
				feedHtmlSpaced.Execute(w, catFeeds[j])
			}
		}
	}
	fmt.Fprintf(w, "<br>")
	//and the categories 
	allFeeds := getFeedsWithoutCats()
	for i := range allFeeds {
		feedHtml.Execute(w, allFeeds[i])
	}
}

//print the list of entries for the selected category, feed, or marked
func handleEntries(w http.ResponseWriter, r *http.Request) {
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// format is /entries/{feed|category|marked}/<id>/{read|unread|marked|next|previous}[/{feed_id|cat_id}]
	var feedOrCat string
	var id string
	var mode string
	var curID string //only really needed for getting the next one in a feed/cat
	pathVars(r, "/entries/", &feedOrCat, &id, &mode, &curID)
	ur := 1    //unread/read to unread by default
	mkd := "0" //marked to unmarked by default
	switch mode {
	case "read":
		ur = 0
		mkd = "%"
	case "marked":
		mkd = "1"
		ur = 0
	case "next":
		var retval string
		if feedOrCat == "feed" {
			stmtNextFeedEntry.QueryRow(curID, id).Scan(&retval)
		} else {
			stmtNextCategoryEntry.QueryRow(curID, id).Scan(&retval)
		}
		fmt.Fprintf(w, retval)
		return
	case "previous":
		var retval string
		if feedOrCat == "feed" {
			stmtPreviousFeedEntry.QueryRow(curID, id).Scan(&retval)
		} else {
			stmtPreviousCategoryEntry.QueryRow(curID, id).Scan(&retval)
		}
		fmt.Fprintf(w, retval)
		return
	}
	//print header for list
	fmt.Fprintf(w, "<form id='entries_form'><table class='headlinesList' id='headlinesList' width='100%'>\n")
	var el []Entry
	switch feedOrCat {
	case "feed":
		el = entriesFromSql(stmtFeedEntries, id, ur, mkd)
	case "category":
		el = entriesFromSql(stmtCatEntries, id, ur, mkd)
	case "marked":
		el = allMarkedEntries()
	}
	if len(el) == 0 {
		fmt.Fprintf(w, "No entries found")
	}
	for a := range el {
		listEntryHtml.Execute(w, el[a])
	}
	//print footer for entries list
	fmt.Fprintf(w, "</form>\n</table>\n")
}

func handleMain(w http.ResponseWriter, r *http.Request) {
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
