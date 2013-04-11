package main

import (
	"html/template"
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
	Expirey        string
	CategoryID     int
	ViewMode       string
	AutoscrollPX   int
	Exclude        string
	ErrorString    string
	ViewModeSelect template.HTML
	CategorySelect template.HTML
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
