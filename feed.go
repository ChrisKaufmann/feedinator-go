package main

import (
	"database/sql"
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
}

func (f Feed) Unread() (count int) {
	mc.GetOr("Feed"+tostr(f.ID)+"_UnreadCount", &count, func() {
		print("fc" + tostr(f.ID) + "-")
		err := stmtFeedUnread.QueryRow(f.ID).Scan(&count)
		if err != nil {
			print(err.Error())
		}
	})
	return count
}
func (f Feed) UnreadEntries() (el []Entry) {
	print("f.unreadentries")
	mc.GetOr("Feed"+tostr(f.ID)+"_unreadentries", &el, func() {
		el = f.GetEntriesByParam("unread = 1")
	})
	return el
}
func (f Feed) MarkedEntries() (el []Entry) {
	print("f.markedentries")
	mc.GetOr("Feed"+tostr(f.ID)+"_markedentries", &el, func() {
		el = f.GetEntriesByParam("marked = 1")
	})
	return el
}
func (f Feed) ReadEntries() (el []Entry) {
	print("f.ReadEntries")
	mc.GetOr("Feed"+tostr(f.ID)+"_readentries", &el, func() {
		el = f.GetEntriesByParam("unread = '0'")
	})
	return el
}
func (f Feed) SearchTitles(s string) (el []Entry) {
	print("f.search("+s+")")
	mc.GetOr("Feed"+tostr(f.ID)+"_search_"+s, &el, func() {
		el = f.GetEntriesByParam("title like '%"+s+"%'")
	})
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
func (f Feed) Next(id string) Entry {
	var e Entry
	nes := "FeedEntry" + id + "Next"
	pes := "FeedEntry" + id + "Previous"
	ce := getEntry(id)
	mc.SetTime(pes, ce, 300)
	err := mc.Get(nes, &e)
	if err != nil {
		var retval string
		stmtNextFeedEntry.QueryRow(tostr(f.ID), id).Scan(&retval)
		e := getEntry(retval)
		mc.SetTime(nes, e, 300)
		return e
	}
	return e
}
func (f Feed) Previous(id string) Entry {
	var retval string
	stmtPreviousFeedEntry.QueryRow(tostr(f.ID), id).Scan(&retval)
	e := getEntry(retval)
	return e
}
func (f Feed) GetEntriesByParam(p string) []Entry {
	var query = "select " + entrySelectString + " from ttrss_entries e where e.feed_id = " + tostr(f.ID) + " and " + p + " order by e.id ASC;"
	el := getEntriesFromSql(query)
	return el
}
func (feed Feed) Print() {
	print("\nFeed:\n" +
		"\tID: " + tostr(feed.ID) +
		"\n\tTitle: " + feed.Title +
		"\n\tURL: " + feed.Url +
		"\n\tUserName: " + feed.UserName +
		"\n\tPublic: " + feed.Public +
		"\n\tCategoryID: " + tostr(feed.CategoryID) +
		"\n\tViewMode: " + feed.ViewMode +
		"\n\tAutoscrollPX: " + tostr(feed.AutoscrollPX) +
		"\n\tExclude: " + feed.Exclude +
		"\n\tExclude Data: " + feed.ExcludeData +
		"\n\tErrorstring: " + feed.ErrorString +
		"\n\tExpirey: " + feed.Expirey +
		"\n\tUnread:" + tostr(feed.Unread()) + "\n")
}

func (f Feed) Save() {
	f.Exclude = html.EscapeString(f.Exclude)
	stmtSaveFeed.Exec(f.Title, f.Url, f.Public, f.CategoryID, f.ViewMode, f.AutoscrollPX, f.Exclude, f.ExcludeData, f.Expirey, f.ID)
	f.ClearCache()
}
func (f Feed) Class() string {
	if f.Unread() > 0 {
		return "oddUnread"
	}
	return "odd"
}
func (f Feed) Update() {
	os.Chdir("update")
	out, err := exec.Command("perl", "update_feeds.pl", "feed_id="+tostr(f.ID)).Output()
	os.Chdir("..")
	fmt.Printf("FeedUpdate: %q\n", out)
	if err != nil {
		err.Error()
	}
	f.ClearCache()

	c := getCat(tostr(f.CategoryID))
	c.ClearCache()

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
func (f Feed) ClearCache() {
	mc.DeleteLike("Feed" + tostr(f.ID) + "_")
	cl := []string{"Feed" + tostr(f.ID) + "_",
		"FeedsWithoutCats" + f.UserName,
		"FeedList",
		"Feed" + tostr(f.ID) + "_UnreadCount",
		"Feed" + tostr(f.ID) + "_readentries",
		"Feed" + tostr(f.ID) + "_unreadentries",
		"Feed" + tostr(f.ID) + "_markedentries",
	}
	for _, i := range cl {
		err := mc.Delete(i)
		if err != nil {
			err.Error()
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
		var query = "delete from ttrss_entries where feed_id=" + tostr(f.ID) + " and title like '%" + e + "%'"
		var stmt = sth(db, query)
		stmt.Exec()
	}
	f.ClearCache()
}
func (f Feed) Insert() {
	if f.Url == "" {
		panic("URL is blank for new feed")
	}
	if f.UserName == "" {
		panic("username is blank fornew feed")
	}
	stmtInsertFeed.Exec(f.Url, f.UserName, f.Title)
	f.ClearCache()
}
func (f Feed) Delete() {
	//first, delete all of the entries that aren't starred
	stmtDeleteFeedEntries.Exec(f.ID)
	//then delete the feed from the feeds table
	stmtDeleteFeed.Exec(f.ID)
	f.ClearCache()
}

var (
	stmtFeedUnread          *sql.Stmt
	stmtGetFeedsWithoutCats *sql.Stmt
	stmtGetFeed             *sql.Stmt
	stmtGetFeeds            *sql.Stmt
	stmtGetAllFeeds         *sql.Stmt
	stmtNextFeedEntry       *sql.Stmt
	stmtPreviousFeedEntry   *sql.Stmt
	stmtSaveFeed            *sql.Stmt
	stmtInsertFeed          *sql.Stmt
	stmtDeleteFeedEntries   *sql.Stmt
	stmtDeleteFeed          *sql.Stmt
)

func init() {
	var selecttxt string = " id, IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''),IFNULL(exclude_data,''), IFNULL(error_string,''), IFNULL(expirey,'') "
	stmtInsertFeed = sth(db, "insert into ttrss_feeds (feed_url,user_name,title) values (?,?,?)")
	stmtGetFeeds = sth(db, "select "+selecttxt+" from ttrss_feeds where user_name = ?")
	stmtGetAllFeeds = sth(db, "select "+selecttxt+"	from ttrss_feeds")
	stmtGetFeed = sth(db, "select "+selecttxt+"	from ttrss_feeds where id = ?")
	stmtFeedUnread = sth(db, "select count(ttrss_entries.id) as unread from ttrss_entries where ttrss_entries.feed_id=? and ttrss_entries.unread='1'")
	stmtGetFeedsWithoutCats = sth(db, "select id from ttrss_feeds where user_name=? and (category_id is NULL or category_id=0) order by id ASC")
	stmtNextFeedEntry = sth(db, "select id from ttrss_entries where feed_id=? and id > ? order by id ASC limit 1")
	stmtPreviousFeedEntry = sth(db, "select id from ttrss_entries where feed_id=? and id<? order by id DESC limit 1")
	stmtSaveFeed = sth(db, "update ttrss_feeds set title=?, feed_url=?,public=?,category_id=?,view_mode=?,autoscroll_px=?,exclude=?,exclude_data=?,expirey=? where id=? limit 1")
	stmtDeleteFeedEntries = sth(db, "delete from ttrss_entries where feed_id=?")
	stmtDeleteFeed = sth(db, "delete from ttrss_feeds where id=? limit 1")
}

func getFeeds() (allFeeds []Feed) {
	mc.GetOr("AllFeeds"+userName+"_", &allFeeds, func() {
		rows, err := stmtGetFeeds.Query(userName)
		if err != nil {
			print(err.Error())
			return
		}
		for rows.Next() {
			var feed Feed
			rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString, &feed.Expirey)
			allFeeds = append(allFeeds, feed)
		}
	})
	return allFeeds
}
func getAllFeeds() []Feed {
	var allFeeds []Feed
	var feedids []int
	err := mc.Get("FeedList", &feedids)
	if err != nil {
		print("-FL<ALL>")
		rows, err := stmtGetAllFeeds.Query()
		if err != nil {
			err.Error()
			return allFeeds
		}
		for rows.Next() {
			var feed Feed
			rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ExcludeData, &feed.ErrorString, &feed.Expirey)
			allFeeds = append(allFeeds, feed)
			feedids = append(feedids, feed.ID)
			mc.Set("Feed"+tostr(feed.ID), feed)
		}
		mc.Set("FeedList", feedids)
	} else {
		print("+FL<ALL>")
		for i := range feedids {
			f := getFeed(tostr(feedids[i]))
			allFeeds = append(allFeeds, f)
		}
	}
	return allFeeds
}
func cacheAllFeeds() {
	_ = getAllFeeds()
}

func getFeedsWithoutCats() (allFeeds []Feed) {
	mc.GetOr("FeedsWithoutCats"+userName, &allFeeds, func() {
		var feedids []int
		print("-")
		rows, err := stmtGetFeedsWithoutCats.Query(userName)
		if err != nil {
			print(err.Error())
		}
		for rows.Next() {
			var id string
			rows.Scan(&id)
			f := getFeed(id)
			allFeeds = append(allFeeds, f)
			feedids = append(feedids, f.ID)
		}
	})
	return allFeeds
}

func getFeed(id string) Feed {
	var feed Feed
	var fcn = "Feed" + id + "_"

	mc.GetOr(fcn, &feed, func() {
		err := stmtGetFeed.QueryRow(id).Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ExcludeData, &feed.ErrorString, &feed.Expirey)
		if err != nil {
			err.Error()
		}
		if feed.Title == "" {
			feed.Title = "--untitled--"
		}
		feed.Title = html.UnescapeString(feed.Title)
	})
	return feed
}
