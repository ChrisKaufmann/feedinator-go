package main

import (
	"./auth"
	"./feed"
	"fmt"
	"github.com/golang/glog"
	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	"github.com/msbranco/goconfig"
	"html"
	"html/template"
	"flag"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	userName          string
    db                        *sql.DB
	cachefile         = "/dev/null"
	indexHtml         = template.Must(template.ParseFiles("templates/index-nologin.html"))
	mainHtml          = template.Must(template.ParseFiles("templates/main.html"))
	categoryHtml      = template.Must(template.ParseFiles("templates/category.html"))
	categoryHtmlS     = template.Must(template.ParseFiles("templates/category_selected.html"))
	feedHtml          = template.Must(template.ParseFiles("templates/feed.html"))
	feedListHtml	  = template.Must(template.ParseFiles("templates/feed_list.html"))
	feedHtmlSpaced    = template.Must(template.ParseFiles("templates/feed_spaced.html"))
	feedMenuHtml      = template.Must(template.ParseFiles("templates/feed_menu.html"))
	catMenuHtml       = template.Must(template.ParseFiles("templates/category_menu.html"))
	entryLinkHtml     = template.Must(template.ParseFiles("templates/entry_link.html"))
	entryHtml         = template.Must(template.ParseFiles("templates/entry.html"))
	menuDropHtml      = template.Must(template.ParseFiles("templates/menu_dropdown.html"))
	categoryPrintHtml = template.Must(template.ParseFiles("templates/category_print.html"))
	entriesListTmpl		=template.Must(template.ParseFiles("templates/entries_list.tmpl"))
	cookieName        string
	viewModes         = [...]string{"Default", "Link", "Extended", "Proxy"}
	port              string
	mc                = easymemcache.New("127.0.0.1:11211")
	environment       string
)

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo"

func init() {
	var err error
	flag.Parse()
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
    db_name, err := c.GetString("DB", "db")
    if err != nil {
        glog.Fatalf("Config: %s",err)
    }
    db_host, err := c.GetString("DB", "host")
    if err != nil {
        glog.Fatalf("Config: %s",err)
    }
    db_user, err := c.GetString("DB", "user")
    if err != nil {
        glog.Fatalf("Config: %s",err)
    }
    db_pass, err := c.GetString("DB", "pass")
    if err != nil {
        glog.Fatalf("Config: %s",err)
    }
    db, err = sql.Open("mysql", db_user+":"+db_pass+"@"+db_host+"/"+db_name)
    if err != nil {
        glog.Fatalf("Config: %s",err)
    }
	feed.Categoryinit(db,&mc)
	feed.Feedinit()
	feed.Entryinit()
	auth.DB(db)
}

func main() {
	defer db.Close()
	http.HandleFunc("/main", handleMain)
	http.HandleFunc("/demo", handleDemo)
	http.HandleFunc("/logout", auth.HandleLogout)
	http.HandleFunc("/authorize", auth.HandleAuthorize)
	http.HandleFunc("/oauth2callback", auth.HandleOAuth2Callback)
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

	go feed.CacheAllCats()  //create cache for categories at startup
	go feed.CacheAllFeeds() //create cache for feeds at startup
	print("Listening on 127.0.0.1:" + port + "\n")
	http.ListenAndServe("127.0.0.1:"+port, nil)
}
func handleCategory(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	var todo string
	var val string
	u.PathVars(r, "/category/", &id, &todo, &val)
	switch todo {
	case "new":
		var c feed.Category
		c.Name = val
		c.Insert(userName)
		fmt.Fprintf(w, "Added")
	case "name":
		c := feed.GetCat(id)
		c.Name = val
		c.Save()
		fmt.Fprintf(w, id+"Renamed: "+val)
	case "desc":
		c := feed.GetCat(id)
		c.Description = val
		c.Save()
		fmt.Fprintf(w, "Desc: "+val)
	case "delete":
		c := feed.GetCat(id)
		c.Delete()
		fmt.Fprintf(w, "Deleted")
	case "update":
		c := feed.GetCat(id)
		c.Update()
		fmt.Fprintf(w, "Updated")
	case "unread":
		c := feed.GetCat(id)
		print("in unread\n")
		fmt.Fprintf(w, strconv.Itoa(c.Unread()))
	case "exclude":
		c := feed.GetCat(id)
		c.Exclude = val
		c.Save()
		fmt.Fprintf(w, "Exclude:"+c.Exclude)
	case "print":
		c := feed.GetCat(id)
		if err := categoryPrintHtml.Execute(w, c); err != nil {
			glog.Errorf("categoryPrintHtml.Execute: %s", err)
		}
	case "clearcache":
		c := feed.GetCat(id)
		c.ClearCache()
	case "deleteexcludes":
		c := feed.GetCat(id)
		c.DeleteExcludes()
	}
	fmt.Printf("handleCategory %v\n", time.Now().Sub(t0))
}
func handleStats(w http.ResponseWriter, r *http.Request) {
	var t0 = time.Now()
	var todo string
	u.PathVars(r, "/stats/", &todo)
	var c string
	switch todo {
	case "entries":
		c, _ = feed.GetEntriesCount()
	}
	fmt.Fprintf(w, c)
	fmt.Printf("handleStats %v\n", time.Now().Sub(t0))
}
func handleNewFeed(w http.ResponseWriter, r *http.Request) {
	var t0 = time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	formurl := r.FormValue("url")
	var f feed.Feed
	f.Url = formurl
	f.UserName = userName
	purl, _ := url.Parse(formurl)
	f.Title = purl.Host
	f.Insert()
	fmt.Fprintf(w, "Added")
	fmt.Printf("handleNewFeed %v\n", time.Now().Sub(t0))
}
func handleFeed(w http.ResponseWriter, r *http.Request) {
	var t0 = time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	var todo string
	var val string
	u.PathVars(r, "/feed/", &id, &todo, &val)
	f := feed.GetFeed(u.Toint(id))
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
		f.AutoscrollPX = u.Toint(val)
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
		f.CategoryID = u.Toint(val)
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
	fmt.Printf("handleFeed %v\n", time.Now().Sub(t0))
}
func handleMarkEntry(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var feedOrCat string
	var fcid string
	var retstr string
	var id string
	var tomark string
	u.PathVars(r, "/entry/mark/",&feedOrCat, &fcid, &id, &tomark)
	print("fc="+feedOrCat+"<=>fcid="+fcid+"<=>id="+id+"<=>tomark="+tomark)
	b := strings.Split(id, ",")
	// one thing can be marked whatever, but a list can only be marked read
	print("\nlength: "+u.Tostr(len(b))+"\n")
	if len(b) == 1 {
		print("Marking: "+tomark)
		retstr = feed.MarkEntry(b[0], tomark,userName)
	} else {
	    switch feedOrCat {
		    case "feed":
				f := feed.GetFeed(u.Toint(fcid))
				err := f.MarkEntriesRead(b)
				if err != nil {
					print(err.Error())
				}
			case "category":
				c := feed.GetCat(fcid)
				err := c.MarkEntriesRead(b)
				if err != nil {
					print(err.Error())
				}
		}
	}
	fmt.Fprintf(w, retstr)
	t1 := time.Now()
	fmt.Printf("handleMarkEntry %v\n", t1.Sub(t0))
}
func handleEntry(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	u.PathVars(r, "/entry/", &id)

	e := feed.GetEntry(id,userName)
	f := feed.GetFeed(e.FeedID)
	e.FeedName = f.Title
	if e.ViewMode() == "link" {
		e.Link = html.UnescapeString(e.Link)
		if err := entryLinkHtml.Execute(w, e); err != nil {
			glog.Errorf("entryLinkHtml: %s", err)
		}
	} else {
		if err := entryHtml.Execute(w, e); err != nil {
			glog.Errorf("entryHtml.Execute: %s", err)
		}
	}
	feed.MarkEntry(id, "read", userName)
	fmt.Printf("handleEntry %v\n", time.Now().Sub(t0))
}
func handleMenu(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var feedOrCat string
	var id string
	var mode string
	var curID string
	var modifier string
	u.PathVars(r, "/menu/", &feedOrCat, &id, &mode, &curID, &modifier)

	switch feedOrCat {
	case "category":
		cat := feed.GetCat(id)
		cat.SearchSelect = getSearchSelect(modifier)
		cat.Search = curID
		if err := catMenuHtml.Execute(w, cat); err != nil {
			glog.Errorf("catMenuHtml: %s", err)
		}
	case "feed":
		f := feed.GetFeed(u.Toint(id))
		f.SearchSelect = getSearchSelect(modifier)
		f.Search = curID
		setSelects(&f)
		if err := feedMenuHtml.Execute(w, f); err != nil {
			glog.Errorf("feedMenuHtml.Execute: %s", err)
		}
	case "marked":
		fmt.Fprintf(w, "&nbsp;")
	}
	fmt.Printf("handleMenu %v\n", time.Now().Sub(t0))
}
func handleSelectMenu(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var id string
	u.PathVars(r, "/menu/select/", &id)
	f := feed.GetFeed(u.Toint(id))
	setSelects(&f)
	if err := menuDropHtml.Execute(w, f); err != nil {
		glog.Errorf("menuDropHtml.Execute: %s", err)
	}
	fmt.Printf("handleSelectMenu %v\n", time.Now().Sub(t0))
}
func getSearchSelect(cur string) template.HTML {
	t0 := time.Now()
	l := []string{"Unread", "Read", "Marked", "All"}
	var h string
	for _, i := range l {
		sel := ""
		if strings.ToLower(i) == strings.ToLower(cur) {
			sel = "selected"
		}
		h = h + "<option value='" + strings.ToLower(i) + "'" + sel + ">" + i + "\n"
	}
	fmt.Printf("handleSearchSelect %v\n", time.Now().Sub(t0))
	return template.HTML(h)
}
func setSelects(f *feed.Feed) {
	t0 := time.Now()
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
	allthecats := feed.GetCategories(userName)
	for i := range allthecats {
		cat := allthecats[i]
		if cat.ID == f.CategoryID {
			cat.Name = "*" + cat.Name
		}
		catHtml = catHtml + "<option value='" + strconv.Itoa(cat.ID) + "'>" + cat.Name + "\n"
	}
	f.ViewModeSelect = template.HTML(optionHtml)
	f.CategorySelect = template.HTML(catHtml)
	fmt.Printf("SetSelects %v\n", time.Now().Sub(t0))
}

//print the list of all feeds
func handleFeedList(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	fmt.Fprintf(w, "<ul class='feedList' id='feedList'>\n")
	allthefeeds := feed.GetFeeds(userName)
	if err := feedListHtml.Execute(w,allthefeeds); err != nil {
		glog.Errorf("feedListHtml.Execute: %s", err)
	}
/*	for i := range allthefeeds {
		f := allthefeeds[i]
		if err := feedHtml.Execute(w, f); err != nil {
			glog.Errorf("feedHtml.Execute: %s", err)
		}
	}
*/
	fmt.Printf("handleFeedList %v\n", time.Now().Sub(t0))
}

//print the list of categories (possibly with feeds in that cat), then the uncategorized feeds
func handleCategoryList(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	fmt.Printf("user: %s", userName)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var currentCat string
	type CategoryList struct {
		SelectedCat	int
		CategoryList []feed.Category
		SelectedFeeds []feed.Feed
		OrphanFeeds	[]feed.Feed
	}
	u.PathVars(r, "/categoryList/", &currentCat)
	fmt.Fprintf(w, "<ul class='feedList' id='feedList'>\n")
	allthecats := feed.GetCategories(userName)
	for _,cat := range allthecats {
		//print the feeds under the currently selected category
		if strconv.Itoa(cat.ID) == currentCat {
			if err := categoryHtmlS.Execute(w, cat); err != nil {
				glog.Errorf("categoryHtmlS.Execute: %s", err)
			}
			fmt.Fprintf(w, "<br>\n")
			catFeeds := cat.Feeds()
			for j := range catFeeds {
				if err := feedHtmlSpaced.Execute(w, catFeeds[j]); err != nil{
					glog.Errorf("feedHtmlSpaced.Execute: %s", err)
				}
			}
		} else {
			if err := categoryHtml.Execute(w, cat); err != nil {
				glog.Errorf("categoryHtml.Execute: %s", err)
			}
			fmt.Fprintf(w, "<br>\n")
		}
	}
	fmt.Fprintf(w, "<hr>")
	allFeeds := feed.GetFeedsWithoutCats(userName)
	for i := range allFeeds {
		if err := feedHtml.Execute(w, allFeeds[i]); err != nil {
			glog.Errorf("feedHtml.Execute: %s", err)
		}
	}
	t1 := time.Now()
	fmt.Printf("handleCategoryList %v\n", t1.Sub(t0))
}

//print the list of entries for the selected category, feed, or marked
func handleEntries(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// format is /entries/{feed|category|marked}/<id>/{read|unread|marked|next|previous}[/{feed_id|cat_id}]
	var feedOrCat string
	var id string
	var mode string
	var curID string    //current entry id for next/previous or search term for search
	var modifier string //secondary mode for next/previous/search (read/unread/marked/etc)
	u.PathVars(r, "/entries/", &feedOrCat, &id, &mode, &curID, &modifier)
	var el []feed.Entry
	switch feedOrCat {
	case "feed":
		f := feed.GetFeed(u.Toint(id))
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
		c := feed.GetCat(id)
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
		el = feed.AllMarkedEntries(userName)
	}
	if len(el) == 0 {
		fmt.Fprintf(w, "No entries found")
	}
	if err := entriesListTmpl.Execute(w,el); err != nil {
		glog.Errorf("entriesListTmpl.Execute: %s", err)
	}
	t1 := time.Now()
	fmt.Printf("handleEntries %v\n", t1.Sub(t0))
}
func handleDemo(w http.ResponseWriter, r *http.Request) {
    t0 := time.Now()
    auth.DemoUser(w,r)
    fmt.Printf("handleDemo %v\n", time.Now().Sub(t0))
}
func handleMain(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := mainHtml.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Printf("handleMain %v\n", time.Now().Sub(t0))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	var loggedin bool
	loggedin, userName = auth.LoggedIn(w, r)
	if !loggedin {
		if err := indexHtml.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if environment == "dev" || environment == "demo" || environment == "test" {
			fmt.Fprintf(w, "<a href='/demo'>Demo</a>")
		}
	} else {
		http.Redirect(w, r, "/main", http.StatusFound)
	}
	fmt.Printf("handleRoot %v\n", time.Now().Sub(t0))
}
