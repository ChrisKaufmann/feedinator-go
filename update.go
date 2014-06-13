package main

import (
	"flag"
)

var (
	userName       string
	cachefile      = "cache.json"
	profileInfoURL string
	cookieName     string
	feed_id		   string
)

func init() {
	flag.StringVar(&feed_id, "feed_id", "all", "Feed Id to update")
	flag.StringVar(&feed_id, "f", "all", "Feed Id to update")
}


func main() {
	// Get cmd line param (if any), and run that one if passed
	flag.Parse()
	if feed_id != "all" {
		print("Updating feed id "+feed_id)
		//update just the one
		f:=getFeed(feed_id)
		print(f.Title)
		f.Update()
	} else {
		print("Updating all feeds")
		//update all feeds
		for _,f := range getAllFeeds() {
			f.Update()
		}
	}
}
