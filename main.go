package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var (
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
)

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo"
const port = "9000"

func init() {
}

func main() {
	http.HandleFunc("/main", handleMain)
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)
	http.HandleFunc("/categoryList/", handleCategoryList)
	http.HandleFunc("/feed/list/", handleFeedList)
	http.HandleFunc("/feed/",handleFeed)
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
func handleFeed(w http.ResponseWriter, r *http.Request) {
	return
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
		feedHtml.Execute(w,allFeeds[i])
	}
	//print the footer for the categories list
	fmt.Fprintf(w, "</ul>\n<td align='right'>\n<form name='add_feed_form'>\n<input type='text' name='add_feed_text'>\n<input type='button' value='Add' onclick='add_feed(this.form)'>\n	</form>\n</td>\n")
}

//print the list of entries for the selected category, feed, or marked
func handleEntries(w http.ResponseWriter, r *http.Request) {
	//var err error
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
	fmt.Fprintf(w,"<form id='entries_form'><table class='headlinesList' id='headlinesList' width='100%'>")
	var el []Entry
	switch feedOrCat {
		case "feed":
			el = entriesFromSql(stmtFeedEntries,id,ur)
		case "category":
			el = entriesFromSql(stmtCatEntries,id,ur)
		case "marked":
			el = entriesFromSql(stmtMarkedEntries,id,ur)
	}
	for a := range el {
		listEntryHtml.Execute(w,el[a])
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


func evenodd(i int) string {
	if i%2 == 0 {
		return "even"
	}
	return "odd"
}
func unescape(s string) string {
	s = strings.Replace(s, "&#34;", "\"", -1)
	s = strings.Replace(s, "&#47;", "/", -1)
	return s
}