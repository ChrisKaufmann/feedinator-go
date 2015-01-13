package main

import (
	"fmt"
	"github.com/ChrisKaufmann/easymemcache"
	"github.com/msbranco/goconfig"
	"html"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	userName          string
	cachefile         = "/dev/null"
	indexHtml         = template.Must(template.ParseFiles("templates/index-nologin.html"))
	mainHtml          = template.Must(template.ParseFiles("templates/main.html"))
	categoryHtml      = template.Must(template.ParseFiles("templates/category.html"))
	categoryHtmlS     = template.Must(template.ParseFiles("templates/category_selected.html"))
	feedHtml          = template.Must(template.ParseFiles("templates/feed.html"))
	feedHtmlSpaced    = template.Must(template.ParseFiles("templates/feed_spaced.html"))
	listEntryHtml     = template.Must(template.ParseFiles("templates/listentry.html"))
	feedMenuHtml      = template.Must(template.ParseFiles("templates/feed_menu.html"))
	catMenuHtml       = template.Must(template.ParseFiles("templates/category_menu.html"))
	entryLinkHtml     = template.Must(template.ParseFiles("templates/entry_link.html"))
	entryHtml         = template.Must(template.ParseFiles("templates/entry.html"))
	menuDropHtml      = template.Must(template.ParseFiles("templates/menu_dropdown.html"))
	categoryPrintHtml = template.Must(template.ParseFiles("templates/category_print.html"))
	cookieName        string
	viewModes         = [...]string{"Default", "Link", "Extended", "Proxy"}
	port              string
	mc                = easymemcache.New("127.0.0.1:11211")
	environment       string
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
	environment, err = c.GetString("Web", "environment")
	if err != nil {
		err.Error()
	}
	cookieName = "feedinator_auth_" + environment
	mc.Prefix = (environment)
}

func main() {
	defer db.Close()
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
	http.HandleFunc("/stats/", handleStats)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/favicon.ico", http.StripPrefix("/favicon.ico", http.FileServer(http.Dir("./static/favicon.ico"))))
	http.HandleFunc("/", handleRoot)

	go cacheAllCats()  //create cache for categories at startup
	go cacheAllFeeds() //create cache for feeds at startup
	print("Listening on 127.0.0.1:" + port + "\n")
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
	case "update":
		c := getCat(id)
		c.Update()
		fmt.Fprintf(w, "Updated")
	case "unread":
		c := getCat(id)
		print("in unread\n")
		fmt.Fprintf(w, strconv.Itoa(c.Unread()))
	case "exclude":
		c := getCat(id)
		c.Exclude = val
		c.Save()
		fmt.Fprintf(w, "Exclude:"+c.Exclude)
	case "print":
		c := getCat(id)
		categoryPrintHtml.Execute(w, c)
	case "clearcache":
		c := getCat(id)
		c.ClearCache()
	case "deleteexcludes":
		c := getCat(id)
		c.DeleteExcludes()
	}
}
func handleStats(w http.ResponseWriter, r *http.Request) {
	var todo string
	pathVars(r, "/stats/", &todo)
	var c string
	switch todo {
	case "entries":
		c, _ = getEntriesCount()
	}
	fmt.Fprintf(w, c)
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
	purl, _ := url.Parse(formurl)
	f.Title = purl.Host
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
		f.Exclude = html.EscapeString(val)
		f.Save()
		fmt.Fprintf(w, "Exclude saved")
	case "excludedata":
		f.ExcludeData = html.EscapeString(val)
		f.Save()
		fmt.Fprintf(w, "Exclude Data Saved")
	case "category":
		f.CategoryID = toint(val)
		f.Save()
		fmt.Fprintf(w, "Category: "+f.Category().Name)
	case "view_mode":
		f.ViewMode = val
		f.Save()
		fmt.Fprintf(w, "View Mode: "+val)
	case "delete":
		f.Delete()
		fmt.Fprintf(w, "Deleted")
	case "update":
		f.Update()
		fmt.Fprintf(w, "Updated")
	case "unread":
		fmt.Fprintf(w, strconv.Itoa(f.Unread()))
	case "deleteexcludes":
		f.DeleteExcludes()
		fmt.Fprintf(w, "Deleted Excludes")
	case "clearcache":
		f.ClearCache()
		fmt.Fprintf(w, "Cleared Cache")
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
		e.Link = html.UnescapeString(e.Link)
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
	var mode string
	var curID string
	var modifier string
	pathVars(r, "/menu/", &feedOrCat, &id, &mode, &curID, &modifier)

	switch feedOrCat {
	case "category":
		cat := getCat(id)
		cat.SearchSelect = getSearchSelect(modifier)
		cat.Search = curID
		catMenuHtml.Execute(w, cat)
	case "feed":
		f := getFeed(id)
		f.SearchSelect = getSearchSelect(modifier)
		f.Search = curID
		feedMenuHtml.Execute(w, f)
	case "marked":
		fmt.Fprintf(w, "&nbsp;")
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
func getSearchSelect(cur string) template.HTML {
	l := []string{"Unread", "Read", "Marked", "All"}
	var h string
	for _, i := range l {
		sel := ""
		if strings.ToLower(i) == strings.ToLower(cur) {
			sel = "selected"
		}
		h = h + "<option value='" + strings.ToLower(i) + "'" + sel + ">" + i + "\n"
	}
	return template.HTML(h)
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
	t0 := time.Now()
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
	t1 := time.Now()
	fmt.Printf("handleFeedList %v\n", t1.Sub(t0))
}

//print the list of categories (possibly with feeds in that cat), then the uncategorized feeds
func handleCategoryList(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
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
		//print the feeds under the currently selected category
		if strconv.Itoa(cat.ID) == currentCat {
			categoryHtmlS.Execute(w, cat)
			fmt.Fprintf(w, "<br>\n")
			catFeeds := cat.Feeds()
			for j := range catFeeds {
				feedHtmlSpaced.Execute(w, catFeeds[j])
			}
		} else {
			categoryHtml.Execute(w, cat)
			fmt.Fprintf(w, "<br>\n")
		}
	}
	fmt.Fprintf(w, "<hr>")
	allFeeds := getFeedsWithoutCats()
	for i := range allFeeds {
		feedHtml.Execute(w, allFeeds[i])
	}
	t1 := time.Now()
	fmt.Printf("handleCategoryList %v\n", t1.Sub(t0))
}

//print the list of entries for the selected category, feed, or marked
func handleEntries(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	if !loggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// format is /entries/{feed|category|marked}/<id>/{read|unread|marked|next|previous}[/{feed_id|cat_id}]
	var feedOrCat string
	var id string
	var mode string
	var curID string    //current entry id for next/previous or search term for search
	var modifier string //secondary mode for next/previous/search (read/unread/marked/etc)
	pathVars(r, "/entries/", &feedOrCat, &id, &mode, &curID, &modifier)
	var el []Entry
	switch feedOrCat {
	case "feed":
		f := getFeed(id)
		switch mode {
		case "read":
			el = f.ReadEntries()
		case "marked":
			el = f.MarkedEntries()
		case "all":
			el = f.AllEntries()
		case "search":
			el = f.SearchTitles(curID, modifier)
		case "next":
			nid := strconv.Itoa(f.Next(curID).ID)
			fmt.Fprintf(w, nid)
			return
		case "previous":
			nid := strconv.Itoa(f.Previous(curID).ID)
			fmt.Fprintf(w, nid)
			return
		default:
			el = f.UnreadEntries()
		}
	case "category":
		c := getCat(id)
		switch mode {
		case "read":
			el = c.ReadEntries()
		case "marked":
			el = c.MarkedEntries()
		case "all":
			el = c.AllEntries()
		case "search":
			el = c.SearchTitles(curID, modifier)
		case "previous":
			nid := strconv.Itoa(c.Previous(curID).ID)
			fmt.Fprintf(w, nid)
			return
		case "next":
			nid := strconv.Itoa(c.Next(curID).ID)
			fmt.Fprintf(w, nid)
			return
		default:
			el = c.UnreadEntries()
		}
	case "marked":
		el = allMarkedEntries()
	}
	//print header for list
	fmt.Fprintf(w, "<form id='entries_form'><table class='headlinesList' id='headlinesList' width='100%'>\n")
	if len(el) == 0 {
		fmt.Fprintf(w, "No entries found")
	}
	for a := range el {
		listEntryHtml.Execute(w, el[a])
	}
	//print footer for entries list
	fmt.Fprintf(w, "</form>\n</table>\n")
	t1 := time.Now()
	fmt.Printf("handleEntries %v\n", t1.Sub(t0))
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
