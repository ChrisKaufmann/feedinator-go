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
	env, err := c.GetString("Web", "environment")
	if err != nil {panic(err.Error())}
	mc.Prefix=(env)
	if feed_id != "all" {
		print("Updating feed id "+feed_id)
		//update just the one
		f:=getFeed(feed_id)
		print(f.Title)
		f.Update()
	} else {
		print("Updating all feeds")
		//update all feeds
		af := getAllFeeds()
		for _,f := range shuffleFeeds(af) {
			f.Update()
		}
	}
}
