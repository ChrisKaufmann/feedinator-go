package main

import (
	"flag"
	"github.com/ChrisKaufmann/easymemcache"
	"github.com/msbranco/goconfig"
)

var (
	userName       string
	cachefile      = "cache.json"
	profileInfoURL string
	cookieName     string
	feed_id		   string
	mc = easymemcache.New("127.0.0.1:11211")
)

func init() {
	flag.StringVar(&feed_id, "feed_id", "all", "Feed Id to update")
	flag.StringVar(&feed_id, "f", "all", "Feed Id to update")
}


func main() {
	// Get cmd line param (if any), and run that one if passed
	flag.Parse()
	c, err := goconfig.ReadConfigFile("config")
	port, err := c.GetString("Web", "port")
	if err != nil {
		err.Error()
	}
	MyUrl, err := c.GetString("Web", "url")
	if err != nil {
		err.Error()
	}
	mc.Prefix=(MyUrl+port)
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
