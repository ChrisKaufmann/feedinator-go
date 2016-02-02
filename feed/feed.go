package feed

import (
	"github.com/golang/glog"
	u "github.com/ChrisKaufmann/goutils"
	"time"
	"database/sql"
	"strconv"
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"html"
	"html/template"
	"os"
	"os/exec"
	"strings"
)

type Feed struct {
	ID             int
	Title          string
	UserName       string
	Evenodd        string
	Url            string
	LastUpdated    string
	Public         string
	CategoryID     int
	ViewMode       string
	AutoscrollPX   int
	Exclude        string
	ExcludeData    string
	ErrorString    string
	Expirey        string
	ViewModeSelect template.HTML
	CategorySelect template.HTML
	SearchSelect   template.HTML
	Search		   string
}
func (f Feed) String() string {
	return fmt.Sprintf("ID: %v, Title: %s, UserName: %s, Evenodd: %s, Url: %s, CategoryID: %v, Unread: %v", f.ID, f.Title, f.UserName, f.Evenodd, f.Url, f.CategoryID, f.Unread())
}

func (f Feed) DecrementUnread() {
	mc.Decrement("Category"+u.Tostr(f.CategoryID)+"_UnreadCount", 1)
	mc.Decrement("Feed"+u.Tostr(f.ID)+"_UnreadCount", 1)
	f.ClearEntries()
	return
}
func (f Feed) IncrementUnread() {
	mc.Increment("Category"+u.Tostr(f.CategoryID)+"_UnreadCount", 1)
	mc.Increment("Feed"+u.Tostr(f.ID)+"_UnreadCount", 1)
	f.ClearEntries()
	return
}
func (f Feed) Unread() (count int) {
	t0 := time.Now()
	mc.GetOr("Feed"+u.Tostr(f.ID)+"_UnreadCount", &count, func() {
		err := stmtFeedUnread.QueryRow(f.ID).Scan(&count)
		if err != nil {
			glog.Errorf("stmtFeedUnread.QueryRow(%s): %s", f.ID, err)
		}
	})
	glog.Infof("feed(%v).Unread(): %v", f.ID, time.Now().Sub(t0))
	return count
}
func (f Feed) UnreadEntries() (el []Entry) {
	print("f.unreadentries")
	mc.GetOr("Feed"+u.Tostr(f.ID)+"_unreadentries", &el, func() {
		el = f.GetEntriesByParam("unread = 1")
	})
	return el
}
func (f Feed) MarkedEntries() (el []Entry) {
	print("f.markedentries")
	mc.GetOr("Feed"+u.Tostr(f.ID)+"_markedentries", &el, func() {
		el = f.GetEntriesByParam("marked = 1")
	})
	return el
}
func (f Feed) ReadEntries() (el []Entry) {
	print("f.ReadEntries")
	mc.GetOr("Feed"+u.Tostr(f.ID)+"_readentries", &el, func() {
		el = f.GetEntriesByParam("unread = '0'")
	})
	return el
}
func (f Feed) SearchTitles(s string,m string) (el []Entry) {
	var ul []Entry
	switch m {
		case "marked":
			ul = f.MarkedEntries()
		case "read":
			ul = f.ReadEntries()
		case "all":
			ul = f.AllEntries()
		default: //default to unread
			ul = f.UnreadEntries()
	}
	if s == "" {return ul}
	if len(ul) != 0 {
		for _,e := range ul {
			if strings.Contains(strings.ToLower(e.Title), strings.ToLower(s)) {
				el = append(el, e)
			}
		}
	}
	return el
}
func (f Feed) AllEntries() (el []Entry) {
	print("f.allEntries")
	el = f.GetEntriesByParam("1=1")
	return el
}
func (f Feed) Excludes() []string {
	return strings.Split(strings.ToLower(f.Exclude), ",")
}
func (f Feed) ExcludesData() []string {
	return strings.Split(strings.ToLower(f.ExcludeData), ",")
}
func (f Feed) Next(id string) (e Entry) {
	var el []Entry
	mc.GetOr("FeedCurrent"+f.UserName, &el, func() {
		el = f.GetEntriesByParam("id > "+id)
	})
	if id == "" {
		return el[0]
	}
	for i,p := range el{
		if u.Tostr(p.ID) == id {
			if i == len(el)-1 {
				return e
			}
			return el[i+1]
		}
	}
	return e
}
func (f Feed) Previous(id string)(e Entry) {
	var el []Entry
	mc.GetOr("FeedCurrent"+f.UserName, &el, func() {
		el = f.GetEntriesByParam("id < "+id)
	})
	if id == "" {
		return el[0]
	}
	for i,p := range el{
		if u.Tostr(p.ID) == id {
			if i == 0 {
				return e
			}
			return el[i-1]
		}
	}
	return e
}
func (f Feed) GetEntriesByParam(p string) []Entry {
	var query = "select " + entrySelectString + " from ttrss_entries e where e.feed_id = " + u.Tostr(f.ID) + " and " + p + " order by e.id ASC;"
	el := getEntriesFromSql(query)
	mc.Set("FeedCurrent"+f.UserName,el)
	return el
}
func (feed Feed) Print() {
	print("\nFeed:\n" +
		"\tID: " + u.Tostr(feed.ID) +
		"\n\tTitle: " + feed.Title +
		"\n\tURL: " + feed.Url +
		"\n\tUserName: " + feed.UserName +
		"\n\tPublic: " + feed.Public +
		"\n\tCategoryID: " + u.Tostr(feed.CategoryID) +
		"\n\tViewMode: " + feed.ViewMode +
		"\n\tAutoscrollPX: " + u.Tostr(feed.AutoscrollPX) +
		"\n\tExclude: " + feed.Exclude +
		"\n\tExclude Data: " + feed.ExcludeData +
		"\n\tErrorstring: " + feed.ErrorString +
		"\n\tExpirey: " + feed.Expirey +
		"\n\tUnread:" + u.Tostr(feed.Unread()) + "\n")
}

func (f Feed) Save() {
	f.Exclude = html.EscapeString(f.Exclude)
	if _,err := stmtSaveFeed.Exec(f.Title, f.Url, f.Public, f.CategoryID, f.ViewMode, f.AutoscrollPX, f.Exclude, f.ExcludeData, f.Expirey, f.ID); err != nil {
		glog.Errorf("stmtSaveFeed.Exec(%s,%s,%s,%s,%s,%s,%s,%s,%s,%v): %s", f.Title, f.Url, f.Public, f.CategoryID, f.ViewMode, f.AutoscrollPX, f.Exclude, f.ExcludeData, f.Expirey, f.ID, err)
	}
	f.ClearCache()
}
func (f Feed) Class() string {
	if f.Unread() > 0 {
		return "oddUnread"
	}
	return "odd"
}
func (f Feed) Category() (c Category) {
	c = GetCat(u.Tostr(f.CategoryID))
	return c
}
func (f Feed) Update() {
	os.Chdir("update")
	out, err := exec.Command("perl", "update_feeds.pl", "feed_id="+u.Tostr(f.ID)).Output()
	if err != nil {
		glog.Errorf("exec.Command(%s,%s,%s): %s","perl", "update_feeds.pl", "feed_id="+u.Tostr(f.ID), err)
		return
	}
	os.Chdir("..")
	fmt.Printf("FeedUpdate: %q\n", out)
	f.ClearCache()

	f.Category().ClearCache()

	// Note actual work is done in makeItemHandler function
	// I just don't like the way this works as much. Personal taste.
	//	feed := rss.New(5, true, chanHandler, makeItemHandler(f))
	//	if err := feed.Fetch(f.Url, nil); err != nil {
	//		err.Error()
	//	}

}
func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	//println(len(newchannels), "new channel(s) in", feed.Url)
	//We're currently ignoring channels
}
/*
func makeItemHandler(f Feed) rss.ItemHandler {
	return func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
		println(len(newitems), "new item(s) in", ch.Title)
		var guid_cache []string
		excludes := f.Excludes()
		for _, i := range newitems {
			// access f as feed here
			guid_cache = append(guid_cache, escape_guid(i.Links[0].Href))
		}
		existing_entries := make(map[string]Entry)
		for _, e := range f.GetEntriesByParam(" guid in ('" + strings.Join(guid_cache, "', '") + "')") {
			existing_entries[e.GUID] = e
		}
		for _, i := range newitems {
			// Check for existing entries in the guid cache and if so, skip
			guid := escape_guid(i.Links[0].Href)
			if _, ok := existing_entries[guid]; ok {
				print(".")
				continue
			}
			var e Entry
			e.Title = html.EscapeString(i.Title)
			skip := false
			for _, ex := range excludes {
				if strings.Contains(strings.ToLower(e.Title), ex) {
					skip = true
					print("s")
					break
				}
			}
			if skip {
				continue
			}
			e.Link = i.Links[0].Href
			e.Date = i.PubDate
			e.Marked = "0"
			e.FeedID = f.ID
			if i.Content != nil {
				e.Content = template.HTML(html.EscapeString(i.Content.Text))
				e.ContentHash = getHash(i.Content.Text)
			} else {
				e.Content = template.HTML(html.EscapeString(i.Description))
				e.ContentHash = getHash(i.Description)
			}
			e.GUID = guid
			e.Unread = true
			e.Normalize()
			e.Save()
			print("+")
		}
		f.ClearCache()
	}
}
*/
func (f Feed) ClearEntries() {
	mc.Delete("Category" + u.Tostr(f.CategoryID) + "_unreadentries")
	mc.Delete("Feed" + u.Tostr(f.ID) + "_unreadentries")
	mc.Delete("Category" + u.Tostr(f.CategoryID) + "_readentries")
	mc.Delete("Feed" + u.Tostr(f.ID) + "_readentries")
}
func (f Feed) ClearMarked() {
	mc.Delete("Feed" + u.Tostr(f.ID) + "_markedentries")
	mc.Delete("Category" + u.Tostr(f.CategoryID) + "_markedentries")
	f.ClearEntries()
}
func (f Feed) ClearCache() {
	mc.DeleteLike("Feed" + u.Tostr(f.ID) + "_")
	cl := []string{"Feed" + u.Tostr(f.ID) + "_",
		"FeedsWithoutCats" + f.UserName,
		"FeedList",
		"Feed" + u.Tostr(f.ID) + "_UnreadCount",
		"Feed" + u.Tostr(f.ID) + "_readentries",
		"Feed" + u.Tostr(f.ID) + "_unreadentries",
		"Feed" + u.Tostr(f.ID) + "_markedentries",
	}
	for _, i := range cl {
		err := mc.Delete(i)
		if err != nil {
			glog.Errorf("mc.Delete(%s): %s",i,err)
			return
		}
	}
}
func (f Feed) DeleteExcludes() {
	el := f.Excludes()
	for _, e := range el {
		e = escape_guid(e)
		if len(e) < 1 {
			continue
		}
		var query = "delete from ttrss_entries where feed_id=" + u.Tostr(f.ID) + " and title like '%" + e + "%'"
		var stmt,err = u.Sth(db, query)
		if err != nil {
			glog.Errorf("u.Sth(db,%s): %s", query, err)
			return
		}
		 if _, err := stmt.Exec(); err != nil {
			glog.Errorf("DeleteExcludes.stmt.Exec: %s", err)
		 }
	}
	ed := f.ExcludesData()
	for _, e := range ed {
		e = escape_guid(e)
		if len(e) < 1 {
			continue
		}
		var query = "delete from ttrss_entries where feed_id=" + u.Tostr(f.ID) + " and content like '%" + e + "%'"
		var stmt,err = u.Sth(db, query)
		if err != nil {
			glog.Errorf("u.Sth(db,%s): %s", query, err)
			return
		}
		if _, err := stmt.Exec(); err != nil {
			glog.Errorf("stmt.Exec: %s", err)
		}
	}
	f.ClearCache()
}
func (f Feed)MarkEntriesRead(ids []string) (err error) {
    if len(ids) == 0 {
        err = fmt.Errorf("Ids is null")
		return err
    } else {
		// make sure they're all integers
		var id_list []string
        for _,i := range(ids) {
            if _, err := strconv.Atoi(i); err == nil {
				id_list = append(id_list, i)
            }
        }
		if len(id_list) < 1 {
			err = fmt.Errorf("Not enough valid ids passed")
			return err
		}
        j := strings.Join(id_list,",")
		sql := "update ttrss_entries set unread=0 where feed_id="+u.Tostr(f.ID)+" and id in ("+j+")"
        stmtUpdateMarkEntries,err := u.Sth(db, sql)
		if err != nil {
			glog.Errorf("u.Sth(db,%s): %s",sql,err)
			return err
		}
        if _, err = stmtUpdateMarkEntries.Exec(); err != nil {
			glog.Errorf("stmtUpdateMarkEntries.Exec(): %s", err)
		}
        mc.Decrement("Category"+u.Tostr(f.CategoryID)+"_UnreadCount", uint64(len(ids)))
        mc.Decrement("Feed"+u.Tostr(f.ID)+"_UnreadCount", uint64(len(ids)))
        mc.Delete("Category" + u.Tostr(f.CategoryID) + "_unreadentries")
        mc.Delete("Feed" + u.Tostr(f.ID) + "_unreadentries")
        mc.Delete("Category" + u.Tostr(f.CategoryID) + "_readentries")
        mc.Delete("Feed" + u.Tostr(f.ID) + "_readentries")
    }
    return err
}
func (f Feed) Insert() {
	if f.Url == "" {
		panic("URL is blank for new feed")
	}
	if f.UserName == "" {
		panic("username is blank fornew feed")
	}
	_,err := stmtInsertFeed.Exec(f.Url, f.UserName, f.Title)
	if err != nil {
		glog.Errorf("stmtInsertFeed.Exec(%s,%s,%s): %s", f.Url, f.UserName, f.Title, err)
	}
	mc.Delete("FeedsWithoutCats"+f.UserName)
	f.ClearCache()
}
func (f Feed) Delete() {
	//first, delete all of the entries that aren't starred
	if _,err := stmtDeleteFeedEntries.Exec(f.ID); err != nil { 
		glog.Errorf("stmtDeleteFeedEntries.Exec(%v): %s", f.ID, err)
	}

	//then delete the feed from the feeds table
	if _,err := stmtDeleteFeed.Exec(f.ID); err != nil {
		glog.Errorf("stmtDeleteFeed.Exec(%v): %s", f.ID, err)
	}
	f.ClearCache()
}

var (
	stmtFeedUnread          *sql.Stmt
	stmtGetFeedsWithoutCats *sql.Stmt
	stmtGetFeed             *sql.Stmt
	stmtGetFeeds            *sql.Stmt
	stmtGetAllFeeds         *sql.Stmt
	stmtSaveFeed            *sql.Stmt
	stmtInsertFeed          *sql.Stmt
	stmtDeleteFeedEntries   *sql.Stmt
	stmtDeleteFeed          *sql.Stmt
	stmtGetFeedByCat		*sql.Stmt
)

func Feedinit() {
	var selecttxt string = " id, IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''),IFNULL(exclude_data,''), IFNULL(error_string,''), IFNULL(expirey,'') "
	var err error
	stmtInsertFeed,err = u.Sth(db, "insert into ttrss_feeds (feed_url,user_name,title) values (?,?,?)")
	if err!=nil{glog.Fatalf("stmtInsertFeed: %s",err);}
	stmtGetFeeds,err = u.Sth(db, "select "+selecttxt+" from ttrss_feeds where user_name = ?")
	if err!=nil{glog.Fatalf("stmtGetFeeds: %s",err);}
	stmtGetAllFeeds,err = u.Sth(db, "select "+selecttxt+"	from ttrss_feeds")
	if err!=nil{glog.Fatalf("stmtGetAllFeeds: %s",err);}
	stmtGetFeed,err = u.Sth(db, "select "+selecttxt+"	from ttrss_feeds where id = ?")
	if err!=nil{glog.Fatalf("stmtGetFeed: %s",err);}
	stmtFeedUnread,err = u.Sth(db, "select count(ttrss_entries.id) as unread from ttrss_entries where ttrss_entries.feed_id=? and ttrss_entries.unread='1'")
	if err!=nil{glog.Fatalf("stmtFeedUnread: %s",err);}
	stmtGetFeedsWithoutCats,err = u.Sth(db, "select "+selecttxt+" from ttrss_feeds where user_name=? and (category_id is NULL or category_id=0) order by id ASC")
	if err!=nil{glog.Fatalf("stmtGetFeedsWithoutCats: %s",err);}
	stmtSaveFeed,err = u.Sth(db, "update ttrss_feeds set title=?, feed_url=?,public=?,category_id=?,view_mode=?,autoscroll_px=?,exclude=?,exclude_data=?,expirey=? where id=? limit 1")
	if err!=nil{glog.Fatalf("stmtSaveFeed: %s",err);}
	stmtDeleteFeedEntries,err = u.Sth(db, "delete from ttrss_entries where feed_id=?")
	if err!=nil{glog.Fatalf("stmtDeleteFeedEntries: %s",err);}
	stmtDeleteFeed,err = u.Sth(db, "delete from ttrss_feeds where id=? limit 1")
	if err!=nil{glog.Fatalf("stmtDeleteFeed: %s",err);}
	stmtGetFeedByCat, err = u.Sth(db, "select "+selecttxt+" from ttrss_feeds where category_id=?")
	if err!=nil{glog.Fatalf("stmtGetFeedByCat: %s", err);}
}

func GetFeeds(userName string) (allFeeds []Feed) {
	mc.GetOr("AllFeeds"+userName+"_", &allFeeds, func() {
		rows, err := stmtGetFeeds.Query(userName)
		if err != nil {
			glog.Errorf("stmtGetFeeds.Query(%s): %s", userName, err)
			return
		}
		for rows.Next() {
			var feed Feed
			rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ExcludeData, &feed.ErrorString, &feed.Expirey)
			allFeeds = append(allFeeds, feed)
			mc.Set("Feed"+u.Tostr(feed.ID), feed)
		}
	})
	return allFeeds
}
func GetCategoryFeeds(cid int) (cf []Feed) {
	mc.GetOr("CategoryFeeds"+u.Tostr(cid)+"_", &cf, func() {
		for _,f := range GetAllFeeds() {
			if f.CategoryID == cid {
				cf = append(cf, f)
			}
		}
	})
	return cf
}
func GetAllFeeds() []Feed {
	var allFeeds []Feed
	var feedids []int
	err := mc.Get("FeedList", &feedids)
	if err != nil { // not cached
		print("-FL<ALL>")
		rows, err := stmtGetAllFeeds.Query()
		if err != nil {
			glog.Errorf("stmtGetAllFeeds.Query(): %s", err)
			return allFeeds
		}
		for rows.Next() {
			var feed Feed
			rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ExcludeData, &feed.ErrorString, &feed.Expirey)
			allFeeds = append(allFeeds, feed)
			feedids = append(feedids, feed.ID)
			mc.Set("Feed"+u.Tostr(feed.ID), feed)
		}
		mc.Set("FeedList", feedids)
	} else {
		print("+FL<ALL>")
		for i := range feedids {
			f := GetFeed(feedids[i])
			allFeeds = append(allFeeds, f)
		}
	}
	return allFeeds
}
func CacheAllFeeds() {
	_ = GetAllFeeds()
	return
}

func GetFeedsWithoutCats(userName string) (allFeeds []Feed) {
	mc.GetOr("FeedsWithoutCats"+userName, &allFeeds, func() {
		print("-")
		rows, err := stmtGetFeedsWithoutCats.Query(userName)
		if err != nil {
			glog.Errorf("stmtGetFeedsWithoutCats.Query(%s): %s", userName, err)
		}
		for rows.Next() {
			var feed Feed
			rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ExcludeData, &feed.ErrorString, &feed.Expirey)
			allFeeds = append(allFeeds, feed)
			mc.Set("Feed"+u.Tostr(feed.ID), feed) //cache the feed, because why not
		}
	})
	return allFeeds
}

func GetFeed(id int) Feed {
	var feed Feed
	var fcn = "Feed" + u.Tostr(id) + "_"
	if id < 1 {
		glog.Errorf("Non valid id passed to GetFeed")
		return feed
	}

	mc.GetOr(fcn, &feed, func() {
		err := stmtGetFeed.QueryRow(id).Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ExcludeData, &feed.ErrorString, &feed.Expirey)
		if err != nil {
			glog.Errorf("stmtGetFeed.QueryRow(%s).Scan(...): %s", id, err)
		}
		if feed.Title == "" {
			feed.Title = "--untitled--"
		}
		feed.Title = html.UnescapeString(feed.Title)
		mc.Set("Feed"+u.Tostr(feed.ID), feed)
	})
	return feed
}

func escape_guid(s string) string {
    var htmlCodes = map[string]string{
        "&#34;":   "\"",
        "&#47;":   "/",
        "&#39;":   "'",
        "&#42;":   "*",
        "&#63;":   "?",
        "&#160;":  " ",
        "&#8216;": "'",
        "&#8220;": "'",
        "&#8221;": "'",
        "&#8211;": "-",
        "&#8230;": "...",
        "&#8594;": "->",
        "&quot;":  "'",
        "&amp;":   "&",
        "&#37;":   "%",
    }
    for k, v := range htmlCodes {
        s = strings.Replace(s, k, v, -1)
    }
    return s
}
