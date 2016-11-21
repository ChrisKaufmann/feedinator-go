package feed

import (
	"database/sql"
	"fmt"
	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"os/exec"
	"testing"
	"github.com/stvp/assert"
)

var (
	tmc  = easymemcache.New("127.0.0.1:11211")
	d_en *sql.Stmt
	p_en *sql.Stmt
	d_fe *sql.Stmt
	p_fe *sql.Stmt
	d_ca *sql.Stmt
	p_ca *sql.Stmt
)

func init() {
	var err error
	db, err = sql.Open("mysql", "feedinator_test:feedinator_test@tcp(localhost:3306)/feedinator_test")
	if err != nil {
		glog.Fatalf("error: %s", err)
	}
	mc = &tmc
	initDB()
	if d_ca, err = u.Sth(db, "delete from ttrss_categories"); err != nil {
		glog.Fatalf("delete from categories: %s", err)
	}
	if p_ca, err = u.Sth(db, "insert into ttrss_categories(id, name, user_name) values (1,'cat1','test'),(2,'cat2','test'),(3,'cat3','test'),(4,'cat4','test'),(5,'cat5','other'),(6,'cat6','other'),(7,'cat7','other'),(8,'cat8','other')"); err != nil {
		glog.Fatalf("seed categories: %s", err)
	}
	if d_fe, err = u.Sth(db, "delete from ttrss_feeds"); err != nil {
		glog.Fatalf("delete from feeds: %s", err)
	}
	if p_fe, err = u.Sth(db, "insert into ttrss_feeds (id,title,feed_url,user_name,category_id,view_mode) values (1,'test1','http://blah','test',1,'default'),(2,'test2','http://blah','test',1,'default'),(3,'test3','http://blah','other',2,'default'),(4,'test4','http://blah','other',NULL,'default'),(5,'feed','http://blah','test',NULL,'default'),(6,'feed2','http://blah','test',NULL,'default'),(7,'feed3','http://blah','other',NULL,'default'),(8,'feed4','http://blah','other',NULL,'default')"); err != nil {
		glog.Fatalf("Seed feeds: %s", err)
	}
	if d_en, err = u.Sth(db, "delete from ttrss_entries"); err != nil {
		glog.Fatalf("delete from entries: %s", err)
	}
	if p_en, err = u.Sth(db, "insert into ttrss_entries(id,feed_id,guid,content_hash,title,user_name,link,content) values (1,1,1,1,'asdf','test','link','content1'),(2,2,2,2,'asdf','test','link','content2'),(3,1,3,3,'another','other','link','content3'),(4,1,4,4,'another','other','link','content4'),(5,5,5,5,'title','yetanother','link','content5')"); err != nil {
		glog.Fatalf("insert into entries: %s", err)
	}
	Categoryinit(db, mc)
	Feedinit()
	Entryinit()
}
func TestCategory_GetAllCategories(t *testing.T) {
	seed()
	assert.Equal(t,8,cl())
	var cn Category
	cn.Name = "newcat"
	cn.Insert("test")
	assert.Equal(t, 9, cl())
	c1 := GetCat("1")
	c1.Delete()
	c2 := GetCat("2")
	c2.Delete()
	assert.Equal(t, 7, cl())
}
func TestCategory_Save(t *testing.T) {
	seed()
	c := GetCat("1")
	c.Name = "NewCat0"
	c.Save()
	d := GetCat("1")
	assert.Equal(t, "NewCat0", d.Name)
}
func TestCategory_Insert(t *testing.T) {
	seed()
	ctl := cl()
	var c Category
	c.Name = "New Category"
	c.Insert("test")
	assert.Equal(t, cl(), ctl+1)
}
func TestCategory_Delete(t *testing.T) {
	seed()
	c := GetCat("1")
	ctl := cl()
	c.Delete()
	assert.Equal(t, cl(), ctl-1)
}
func TestCategory_Unread(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 4, c1.Unread())
}
func TestCategory_Class(t *testing.T) {
	seed()
	assert.Equal(t, "oddUnread", GetCat("1").Class())
	assert.Equal(t, "odd", GetCat("2").Class())
}
func TestCategory_Excludes(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 0, len(c1.Excludes()))
	c1.Exclude = "a,b,c,d,"
	c1.Save()
	assert.Equal(t, 4,  len(c1.Excludes()))
}
func TestCategory_DeleteExcludes(t *testing.T) {
	seed()
	c1 := GetCat("1")
	c1.DeleteExcludes()
	assert.Equal(t, 4, c1.Unread())
	c1.Exclude = "asdf"
	c1.DeleteExcludes()
	assert.Equal(t, 2, c1.Unread())
}
func TestCategory_SearchTitles(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 2, len(c1.SearchTitles("asdf", "unread")))
	assert.Equal(t, 2, len(c1.SearchTitles("asdf", "all")))
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c1.SearchTitles("asdf", "read")))
	err = e1.Mark()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c1.SearchTitles("asdf", "marked")))
}
func TestCategory_MarkedEntries(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 0, len(c1.MarkedEntries()))
	e1 := GetEntry("1", "test")
	err := e1.Mark()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c1.MarkedEntries()))
}
func TestCategory_UnreadEntries(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 4, len(c1.UnreadEntries()))
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(c1.UnreadEntries()))
}
func TestCategory_ReadEntries(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 0, len(c1.ReadEntries()))
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(c1.ReadEntries()))
}
func TestCategory_AllEntries(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 4, len(c1.AllEntries()))
}
func TestCategory_GetEntriesByParam(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 4, len(c1.GetEntriesByParam("1=1")))
}
func TestGetCat(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t,1, c1.ID)
}
func TestCategory_Feeds(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 2, len(c1.Feeds()))
	f4, err := GetFeed(4)
	assert.Nil(t, err)
	f4.CategoryID = 1
	f4.Save()
	assert.Equal(t, 3, len(c1.Feeds()))
}
func TestCategory_FeedsStr(t *testing.T) {
	seed()
	c1 := GetCat("1")
	assert.Equal(t, 2, len(c1.FeedsStr()))
	f4, err := GetFeed(4)
	assert.Nil(t, err)
	f4.CategoryID = 1
	f4.Save()
	assert.Equal(t, 3, len(c1.FeedsStr()))
}
func TestGetCategories(t *testing.T) {
	seed()
	assert.Equal(t, 4, len(GetCategories("test")))
	var newcat Category
	newcat.Name = "newest!"
	newcat.UserName = "test"
	newcat.Insert("test")
	assert.Equal(t, 5, len(GetCategories("test")))
}
func TestCategory_MarkEntriesRead(t *testing.T) {
	seed()
	c1 := GetCat("1")
	el := []string{"1", "2"}
	err := c1.MarkEntriesRead(el)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(c1.UnreadEntries()))
}
func TestCacheAllCats(t *testing.T) {
	seed()
	var c Category
	var cc int
	err := mc.Get("Category1_", &c)
	if err == nil {
		t.Errorf("mc.Get(Category1_) did not error")
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err == nil {
		t.Errorf("mc.Get(Category1_UnreadCount) did not error")
	}
	CacheAllCats()
	err = mc.Get("Category1_", &c)
	if err != nil {
		t.Errorf("mc.Get(Category1_): %s", err)
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err != nil {
		t.Errorf("mc.Get(Category1_UnreadCount): %s", err)
	}
}
func TestCategory_ClearCache(t *testing.T) {
	seed()
	var c Category
	var cc int
	err := mc.Get("Category1_", &c)
	if err == nil {
		t.Errorf("mc.Get(Category1_) did not error")
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err == nil {
		t.Errorf("mc.Get(Category1_UnreadCount) did not error")
	}
	CacheAllCats()
	err = mc.Get("Category1_", &c)
	if err != nil {
		t.Errorf("mc.Get(Category1_): %s", err)
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err != nil {
		t.Errorf("mc.Get(Category1_UnreadCount): %s", err)
	}
	c.ClearCache()
	err = mc.Get("Category1_", &c)
	if err == nil {
		t.Errorf("mc.Get(Category1_) did not error")
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	assert.NotNil(t, err)
	if err == nil {
		t.Errorf("mc.Get(Category1_UnreadCount) did not error")
	}

}
func TestGetAllCategories(t *testing.T) {
	seed()
	assert.Equal(t, 8, len(GetAllCategories()))
}

func cl() int {
	return len(GetAllCategories())
}

func initDB() {
	if _, err := db.Query("drop table if exists ttrss_categories"); err != nil {
		glog.Errorf("Drop table categories: %s", err)
	}
	if _, err := db.Query("CREATE TABLE `ttrss_categories` (  `name` varchar(64) DEFAULT NULL,  `user_name` varchar(64) DEFAULT NULL,  `description` varchar(255) DEFAULT NULL,  `id` int(11) NOT NULL AUTO_INCREMENT,  `order_num` int(11) DEFAULT NULL,  `exclude` text,  PRIMARY KEY (`id`),  KEY `idx_user_id` (`user_name`,`id`)) ENGINE=MyISAM"); err != nil {
		glog.Fatalf("Creat table categories: %s", err)
	}
	if _, err := db.Query("drop table if exists ttrss_entries"); err != nil {
		glog.Errorf("Drop table entries: %s", err)
	}
	if _, err := db.Query("CREATE TABLE `ttrss_entries` (  `id` int(11) NOT NULL AUTO_INCREMENT,  `feed_id` int(11) NOT NULL DEFAULT '0',  `updated` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',  `title` text NOT NULL,  `guid` varchar(255) NOT NULL DEFAULT '',  `link` text NOT NULL,  `content` text NOT NULL,  `content_hash` varchar(250) NOT NULL DEFAULT '',  `last_read` datetime DEFAULT NULL,  `marked` tinyint(1) NOT NULL DEFAULT '0',  `date_entered` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',  `no_orig_date` tinyint(1) NOT NULL DEFAULT '0',  `comments` varchar(250) NOT NULL DEFAULT '',  `unread` enum('1','0') NOT NULL DEFAULT '1',  `extended_content` text,  `user_name` varchar(128) DEFAULT NULL,  PRIMARY KEY (`id`),  KEY `feed_id` (`feed_id`),  KEY `idx_user_guid` (`user_name`,`guid`),  KEY `idx_entries_marked_username_date` (`marked`,`user_name`,`date_entered`),  KEY `idx_entries_feedid_entered` (`feed_id`,`date_entered`),  KEY `idx_feedid_unread` (`feed_id`,`unread`),  KEY `feed_id_guid` (`feed_id`,`guid`(35)))"); err != nil {
		glog.Fatalf("Create table entries: %s", err)
	}
	if _, err := db.Query("drop table if exists ttrss_feeds"); err != nil {
		glog.Errorf("Drop table feeds: %s", err)
	}
	if _, err := db.Query("CREATE TABLE `ttrss_feeds` (  `id` int(11) NOT NULL AUTO_INCREMENT,  `title` varchar(200) NOT NULL DEFAULT '',  `feed_url` varchar(250) NOT NULL DEFAULT '',  `icon_url` varchar(250) DEFAULT NULL,  `last_updated` datetime DEFAULT '0000-00-00 00:00:00',  `user_name` varchar(128) DEFAULT NULL,  `show_extended` tinyint(1) DEFAULT '0',  `public` tinyint(1) DEFAULT '0',  `expirey` text,  `category_id` int(10) unsigned DEFAULT NULL,  `view_mode` enum('default','link','proxy','extended','linknew','proxynew') DEFAULT NULL,  `autoscroll_px` int(10) unsigned DEFAULT NULL,  `category_order_num` int(11) DEFAULT NULL,  `feed_order_num` int(11) DEFAULT NULL,  `exclude` varchar(1024) DEFAULT NULL,  `error_string` varchar(1024) DEFAULT NULL,  `exclude_data` text,  PRIMARY KEY (`id`),  KEY `ttrss_user_id` (`user_name`,`id`),  KEY `idx_category` (`category_id`)) ENGINE=InnoDB"); err != nil {
		glog.Fatalf("create table feeds: %s", err)
	}

	if err := mc.DeleteAll(); err != nil {
		glog.Errorf("Could not clear memcache")
	}
}
func seed() {
	if _, err := d_ca.Exec(); err != nil {
		glog.Fatalf("delete from categories: %s", err)
	}
	if _, err := p_ca.Exec(); err != nil {
		glog.Fatalf("seed categories: %s", err)
	}
	if _, err := d_fe.Exec(); err != nil {
		glog.Fatalf("delete from feeds: %s", err)
	}
	if _, err := p_fe.Exec(); err != nil {
		glog.Fatalf("Seed feeds: %s", err)
	}
	if _, err := d_en.Exec(); err != nil {
		glog.Fatalf("delete from entries: %s", err)
	}
	if _, err := p_en.Exec(); err != nil {
		glog.Fatalf("Seed entries: %s", err)
	}
	if err := mc.DeleteAll(); err != nil {
		glog.Errorf("Could not clear memcache")
	}
}
func dumptable(t string) {
	out, err := exec.Command("mysqldump", "-uroot", "feedinator_test", t).Output()
	if err != nil {
		glog.Errorf("%s", err)
	}
	fmt.Printf("----------------------\n%s----------------------\n", out)
}
