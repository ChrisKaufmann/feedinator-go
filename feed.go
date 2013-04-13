package main

import (
	"database/sql"
	"html/template"
	"strconv"
)

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
	CategoryID     int
	ViewMode       string
	AutoscrollPX   int
	Exclude        string
	ErrorString    string
	Expirey		   string
	ViewModeSelect template.HTML
	CategorySelect template.HTML
}
func (feed Feed) Print() {
	print("Getting feed:\n"+"\tID: "+strconv.Itoa(feed.ID)+"\n\tTitle: "+feed.Title+"\n\tURL: "+feed.Url+"\n\tUserName: "+feed.UserName+"\n\tPublic: "+feed.Public+"\n\tCategoryID: "+strconv.Itoa(feed.CategoryID)+"\n\tViewMode: "+feed.ViewMode+"\n\tAutoscrollPX: "+strconv.Itoa(feed.AutoscrollPX)+"\n\tExclude: "+feed.Exclude+"\n\tErrorstring: "+feed.ErrorString)
}

func (f Feed) Save() {
	stmtSaveFeed.Exec(f.Title,f.Url,f.Public,f.CategoryID,f.ViewMode,f.AutoscrollPX,f.Exclude,f.Expirey,f.ID)
}

var (
	stmtFeedUnread          *sql.Stmt
	stmtGetFeedsWithoutCats *sql.Stmt
	stmtFeedEntries         *sql.Stmt
	stmtGetFeed             *sql.Stmt
	stmtGetFeeds            *sql.Stmt
	stmtNextFeedEntry       *sql.Stmt
	stmtPreviousFeedEntry   *sql.Stmt
	stmtSaveFeed			*sql.Stmt
)

func init() {
	stmtGetFeeds=sth(db,"select id, IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''), IFNULL(error_string,'') from ttrss_feeds where user_name = ?")
	stmtGetFeed=sth(db,"select id,IFNULL(title,''), IFNULL(feed_url,''), IFNULL(last_updated,''), IFNULL(user_name,''), IFNULL(public,''),  IFNULL(category_id,0), IFNULL(view_mode,''), IFNULL(autoscroll_px,0), IFNULL(exclude,''), IFNULL(error_string,''),IFNULL(expirey,'') from ttrss_feeds where id = ?")
	stmtFeedEntries=sth(db,"select e.id from ttrss_entries as e, ttrss_feeds as f where e.feed_id=f.id and f.id = ? and unread= ?")
	stmtFeedUnread=sth(db,"select count(ttrss_entries.id) as unread from ttrss_entries where ttrss_entries.feed_id=? and ttrss_entries.unread='1'")
	stmtGetFeedsWithoutCats=sth(db,"select id from ttrss_feeds where user_name=? and category_id is NULL or category_id=0 order by id ASC")
	stmtNextFeedEntry=sth(db,"select id from ttrss_entries where feed_id=? and id > ? limit 1")
	stmtPreviousFeedEntry=sth(db,"select id from ttrss_entries where feed_id=? and id<? order by id DESC limit 1")
	stmtSaveFeed=sth(db,"update ttrss_feeds set title=?, feed_url=?,public=?,category_id=?,view_mode=?,autoscroll_px=?,exclude=?,expirey=? where id=? limit 1")
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
		rows.Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public,  &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString)
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

func getFeedsWithoutCats() []Feed {
	var allFeeds []Feed
	rows, err := stmtGetFeedsWithoutCats.Query(userName)
	if err != nil {
		err.Error()
	}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		f := getFeed(id)
		allFeeds = append(allFeeds, f)
	}
	return allFeeds
}

func unreadFeedCount(id int) int {
	var count int
	err := stmtFeedUnread.QueryRow(id).Scan(&count)
	if err != nil {
		err.Error()
	}
	return count
}
func getFeed(id string) Feed {
	var feed Feed
	err := stmtGetFeed.QueryRow(id).Scan(&feed.ID, &feed.Title, &feed.Url, &feed.LastUpdated, &feed.UserName, &feed.Public,  &feed.CategoryID, &feed.ViewMode, &feed.AutoscrollPX, &feed.Exclude, &feed.ErrorString,&feed.Expirey)
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
