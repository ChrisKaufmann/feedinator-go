package main

import (
	"../feed"
	"database/sql"
	"flag"
	"github.com/ChrisKaufmann/easymemcache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/msbranco/goconfig"
)

var (
	userName       string
	cachefile      = "cache.json"
	profileInfoURL string
	cookieName     string
	feed_id        int
	environment    string
	config_file    string
	mc             = easymemcache.New("127.0.0.1:11211")
	db             *sql.DB
)

func init() {
	var err error
	flag.IntVar(&feed_id, "feed_id", 0, "Feed Id to update")
	flag.IntVar(&feed_id, "f", 0, "Feed Id to update")
	flag.StringVar(&config_file, "config", "../config", "Config file")
	flag.StringVar(&config_file, "c", "../config", "Config file")

	c, err := goconfig.ReadConfigFile(config_file)
	if err != nil {
		glog.Fatalf("Couldn't parse config file %s: %s", config_file, err)
	}
	environment, err = c.GetString("Web", "environment")
	if err != nil {
		glog.Fatalf("Config: %s", err)
	}
	db_name, err := c.GetString("DB", "db")
	if err != nil {
		glog.Fatalf("Config: %s", err)
	}
	db_host, err := c.GetString("DB", "host")
	if err != nil {
		glog.Fatalf("Config: %s", err)
	}
	db_user, err := c.GetString("DB", "user")
	if err != nil {
		glog.Fatalf("Config: %s", err)
	}
	db_pass, err := c.GetString("DB", "pass")
	if err != nil {
		glog.Fatalf("Config: %s", err)
	}
	db, err = sql.Open("mysql", db_user+":"+db_pass+"@"+db_host+"/"+db_name)
	if err != nil {
		glog.Fatalf("Config: %s", err)
	}
	feed.Categoryinit(db, &mc)
	feed.Feedinit()
	feed.Entryinit()
	mc.Prefix = (environment)
}

func main() {
	// Get cmd line param (if any), and run that one if passed
	flag.Parse()
	if feed_id != 0 {
		glog.Infof("Updating feed id %v", feed_id)
		//update just the one
		f, err := feed.GetFeed(feed_id)
		if err != nil {
			glog.Errorf("feed.GetFeed(%v): %s", feed_id, err)
			return
		}
		err = f.Update()
		if err != nil {
			glog.Errorf("feed(%v).Update(): %s", f.ID, err)
		}
	} else {
		print("Updating all feeds")
		//fire up a few backend processes to handle
		c := make(chan feed.Feed)
		for i := 0; i < 10; i++ {
			go PollChannel(c)
		}
		//update all feeds
		af := feed.GetAllFeeds()
		for _, f := range af {
			c <- f
		}
	}
}
func PollChannel(queue chan feed.Feed) {
	print("Starting PollChannel\n")
	for f := range queue {
		err := f.Update()
		if err != nil {
			glog.Errorf("feed(%v).Update(): %s", f.ID, err)
		}
	}
}
