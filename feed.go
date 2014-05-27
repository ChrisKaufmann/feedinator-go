package main

import (
	"os"
	"fmt"
	"os/exec"
	"database/sql"
	"html"
	"html/template"
	"strconv"
	"encoding/json"
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
	ErrorString    string
	Expirey        string
	ViewModeSelect template.HTML
	CategorySelect template.HTML
}

func (f Feed) Unread() int {
	var count int
	err := stmtFeedUnread.QueryRow(f.ID).Scan(&count)
	if err != nil {
		err.Error()
	}
	return count
}
func (feed Feed) Print() {
	print("Getting feed:\n" + "\tID: " + strconv.Itoa(feed.ID) + "\n\tTitle: " + feed.Title + "\n\tURL: " + feed.Url + "\n\tUserName: " + feed.UserName + "\n\tPublic: " + feed.Public + "\n\tCategoryID: " + strconv.Itoa(feed.CategoryID) + "\n\tViewMode: " + feed.ViewMode + "\n\tAutoscrollPX: " + strconv.Itoa(feed.AutoscrollPX) + "\n\tExclude: " + feed.Exclude + "\n\tErrorstring: " + feed.ErrorString)
}

func (f Feed) Save() {
	stmtSaveFeed.Exec(f.Title, f.Url, f.Public, f.CategoryID, f.ViewMode, f.AutoscrollPX, f.Exclude, f.Expirey, f.ID)
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
	out,err := exec.Command("perl","update_feeds.pl","feed_id="+strconv.Itoa(f.ID)).Output()
	os.Chdir("..")
	fmt.Printf("FeedUpdate: %q\n", out)
	if err != nil {
		err.Error()
	}
	f.ClearCache()
	c := getCat(strconv.Itoa(f.CategoryID))
	c.ClearCache()
}
func (f Feed) ClearCache() {
	err := mc.Delete("Feed"+strconv.Itoa(f.ID))
	if err != nil {
		err.Error()
	}
	err = mc.Delete("FeedUnreadCount"+strconv.Itoa(f.ID))
	if err != nil {
		err.Error()
	}
}
func (f Feed) Insert() {
	if f.Url == "" {
		panic("URL is blank for new feed")
	}
	if f.UserName == "" {
		panic("username is blank fornew feed")
	}
	stmtInsertFeed.Exec(f.Url, f.UserName, f.Title)
}
func (f Feed) Delete() {
	f.ClearCache()
	//first, delete all of the entries that aren't starred
	stmtDeleteFeedEntries.Exec(f.ID)

	//then delete the feed from the feeds table
	stmtDeleteFeed.Exec(f.ID)
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
	stmtInsertFeed = sth(db, "insert into ttrss_feeds (feed_url,user_name,title) values (?,?,?)")
	stmtGetFeeds = sth(db, "select id, IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''), IFNULL(error_string,'') from ttrss_feeds where user_name = ?")
	stmtGetAllFeeds = sth(db, "select id, IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''), IFNULL(error_string,'') from ttrss_feeds")
	stmtGetFeed = sth(db, "select id,IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''), IFNULL(error_string,''),IFNULL(expirey,'') from ttrss_feeds where id = ?")
	stmtFeedUnread = sth(db, "select count(ttrss_entries.id) as unread from ttrss_entries where ttrss_entries.feed_id=? and ttrss_entries.unread='1'")
	stmtGetFeedsWithoutCats = sth(db, "select id from ttrss_feeds where user_name=? and (category_id is NULL or category_id=0) order by id ASC")
	stmtNextFeedEntry = sth(db, "select id from ttrss_entries where feed_id=? and id > ? limit 1")
	stmtPreviousFeedEntry = sth(db, "select id from ttrss_entries where feed_id=? and id<? order by id DESC limit 1")
	stmtSaveFeed = sth(db, "update ttrss_feeds set title=?, feed_url=?,public=?,category_id=?,view_mode=?,autoscroll_px=?,exclude=?,expirey=? where id=? limit 1")
	stmtDeleteFeedEntries = sth(db, "delete from ttrss_entries where feed_id=?")
	stmtDeleteFeed = sth(db, "delete from ttrss_feeds where id=? limit 1")
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
		rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString)
		allFeeds = append(allFeeds, feed)
	}
	return allFeeds
}
func getAllFeeds() []Feed {
	var allFeeds []Feed
	var feedids []int
	fl, err := mc.Get("FeedList")
	if err != nil {
		print("-FL<ALL>")
		rows,err := stmtGetAllFeeds.Query()
		if err != nil {
			err.Error()
			return allFeeds
		}
		for rows.Next(){
			var feed Feed
			rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString)
			allFeeds = append(allFeeds,feed)
			feedids = append(feedids, feed.ID)
			mcset("Feed"+strconv.Itoa(feed.ID), feed)
		}
		mcset("FeedList", feedids)
	} else {
		print("+FL<ALL>")
		err = json.Unmarshal(fl.Value, &feedids)
		for i := range feedids {
			f := getFeed(strconv.Itoa(feedids[i]))
			allFeeds = append(allFeeds,f)
		}
	}
	return allFeeds
}
func cacheAllFeeds() {
	_ = getAllFeeds()
}

func getFeedsWithoutCats() []Feed {
	var allFeeds []Feed
	var feedids  []int
	var fcn = "FeedsWithoutCats" + userName
	fwc, err := mc.Get(fcn)
	if err != nil {
		print("-"+fcn)
		rows, err := stmtGetFeedsWithoutCats.Query(userName)
		if err != nil {
			err.Error()
		}
		for rows.Next() {
			var id string
			rows.Scan(&id)
			f := getFeed(id)
			allFeeds = append(allFeeds, f)
			feedids = append(feedids, f.ID)
		}
		mcset(fcn, feedids)
	} else {
		print("+"+fcn)
		err = json.Unmarshal(fwc.Value, &feedids)
		for i := range feedids {
			f := getFeed(strconv.Itoa(feedids[i]))
			allFeeds = append(allFeeds,f)
		}
	}
	return allFeeds
}

func getFeed(id string) Feed {
	var feed Feed
	var fcn = "Feed_" + id

	nufeed, err := mc.Get(fcn)
	if err != nil { //cache miss
		err := stmtGetFeed.QueryRow(id).Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public, &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString, &feed.Expirey)
		if err != nil {
			err.Error()
		}
		if feed.Title == "" {
			feed.Title = "--untitled--"
		}
		feed.Title = html.UnescapeString(feed.Title)
		print("-feed" + id)
		mcset(fcn, feed)
	} else {
		err = json.Unmarshal(nufeed.Value, &feed)
		print("+feed" + id)
	}
	return feed
}
