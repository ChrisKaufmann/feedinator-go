package main

import (
	"html/template"
	"strconv"
	"database/sql"
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
func entriesFromSql(s *sql.Stmt,id string, ur int) []Entry {
	rows,err := s.Query(id,ur)
	var el []Entry
	if err != nil {
		err.Error()
	}
	var count int
	for rows.Next() {
		var id string
		rows.Scan(&id)
		e:=getEntry(id)
		e.Evenodd=evenodd(count)
		el = append(el,e)
		count=count+1
	}
	return el
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
func getFeedsWithoutCats() []Feed {
	var allFeeds []Feed
	rows, err := stmtGetFeedsWithoutCats.Query(userName)
	if err != nil {
		err.Error()
	}
	for rows.Next(){
		var id string
		rows.Scan(&id)
		f :=getFeed(id)
		allFeeds=append(allFeeds,f)
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
func toint(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
