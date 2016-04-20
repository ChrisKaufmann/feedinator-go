package feed

import (
	"database/sql"
	"fmt"
	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"testing"
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
	if p_fe, err = u.Sth(db, "insert into ttrss_feeds (id,title,feed_url,user_name,category_id) values (1,'test1','http://blah','test',1),(2,'test2','http://blah','test',1),(3,'test3','http://blah','other',2),(4,'test4','http://blah','other',NULL),(5,'feed','http://blah','test',NULL),(6,'feed2','http://blah','test',NULL),(7,'feed3','http://blah','other',NULL),(8,'feed4','http://blah','other',NULL)"); err != nil {
		glog.Fatalf("Seed feeds: %s", err)
	}
	if d_en, err = u.Sth(db, "delete from ttrss_entries"); err != nil {
		glog.Fatalf("delete from entries: %s", err)
	}
	if p_en, err = u.Sth(db, "insert into ttrss_entries(id,feed_id,guid,content_hash,title,user_name,link) values (1,1,1,1,'asdf','test','link'),(2,2,2,2,'asdf','test','link'),(3,1,3,3,NULL,'other','link'),(4,1,4,4,NULL,'other','link'),(5,5,5,5,NULL,'yetanother','link')"); err != nil {
		glog.Fatalf("insert into entries: %s", err)
	}
	Categoryinit(db, mc)
	Feedinit()
	Entryinit()
}
func TestCategory_GetAllCategories(t *testing.T) {
	fmt.Printf("\tCategory.GetAllCategories\n")
	seed()
	print("\t\tInitial\n")
	el := cl()
	if el != 8 {
		t.Errorf("GetAllCategories expected len 8, got %v", el)
	}
	print("\t\tAdding\n")
	var cn Category
	cn.Name = "newcat"
	cn.Insert("test")
	il := cl()
	if il != 9 {
		t.Errorf("GetAllCategories() len 9 <=> %v", il)
	}
	print("\t\tDeleting\n")
	c1 := GetCat("1")
	c1.Delete()
	c2 := GetCat("2")
	c2.Delete()
	dl := cl()
	if dl != 7 {
		t.Errorf("GetAllCategories() len 7 <=> %v", dl)
	}
}
func TestCategory_Save(t *testing.T) {
	print("\tCategory.Save()\n")
	seed()
	c := GetCat("1")
	c.Name = "NewCat0"
	c.Save()
	d := GetCat("1")
	if d.Name != "NewCat0" {
		t.Errorf("Cat save didn't work, expected NewCat0, got %s", d.Name)
	}
}
func TestCategory_Insert(t *testing.T) {
	print("\tCategory.Insert()\n")
	seed()
	ctl := cl()
	var c Category
	c.Name = "New Category"
	c.Insert("test")
	if cl() != ctl+1 {
		t.Errorf("Length of category list did not increase")
	}
}
func TestCategory_Delete(t *testing.T) {
	print("\tCategory.Delete()\n")
	seed()
	c := GetCat("1")
	ctl := cl()
	c.Delete()
	if cl() != ctl-1 {
		t.Errorf("Length of GetAllCategories did not decrease")
	}
}
func TestCategory_Unread(t *testing.T) {
	print("\tCategory.Unread()\n")
	seed()
	c1 := GetCat("1")
	if c1.Unread() != 4 {
		t.Errorf("Category Unread 4 <=> %v", c1.Unread())
	}
}
func TestCategory_Class(t *testing.T) {
	print("\tCategory.Class\n")
	print("\t\tUnread\n")
	seed()
	c1 := GetCat("1")
	if c1.Class() != "oddUnread" {
		t.Errorf("Category1.Class not oddUnread: %s", c1.Class())
	}
	print("\t\tRead\n")
	c2 := GetCat("2")
	if c2.Class() != "odd" {
		t.Errorf("Category2.Class not odd: %s", c2.Class())
	}
}
func TestCategory_Excludes(t *testing.T) {
	print("\tCategory.Excludes\n")
	seed()
	c1 := GetCat("1")
	if len(c1.Excludes()) != 0 {
		t.Errorf("c1.Excludes len 0 <=> %v", len(c1.Excludes()))
	}
	c1.Exclude = "a,b,c,d,"
	c1.Save()
	if len(c1.Excludes()) != 4 {
		t.Errorf("C1.Excludes len: 4 <=> %v", len(c1.Excludes()))
	}
}
func TestCategory_DeleteExcludes(t *testing.T) {
	print("\tCategory.DeleteExcludes\n")
	seed()
	c1 := GetCat("1")
	c1.DeleteExcludes()
	print("\t\tDon't delete any if excludes is blank\n")
	if c1.Unread() != 4 {
		t.Errorf("DeleteExcludes unread: 4 <=> %v", c1.Unread())
	}
	print("\t\tDelete when excludes is populated\n")
	c1.Exclude = "asdf"
	c1.DeleteExcludes()
	if c1.Unread() != 2 {
		t.Errorf("DeleteExcludes after delete: 2 <=> %v", c1.Unread())
	}
}
func TestCategory_SearchTitles(t *testing.T) {
	print("\tCategory.SearchTitles\n")
	seed()
	c1 := GetCat("1")
	print("\t\tUnread\n")
	sl := len(c1.SearchTitles("asdf", "unread"))
	if sl != 2 {
		t.Errorf("category.Search(asdf,unread) len: 2 <=> %v", sl)
	}
	print("\t\tAll\n")
	al := len(c1.SearchTitles("asdf", "all"))
	if al != 2 {
		t.Errorf("category.Search(asdf,all) len: 2 <=> %v", al)
	}
	print("\t\tRead\n")
	e1 := GetEntry("1", "test")
	if err := e1.MarkRead(); err != nil {
		t.Errorf("entry(1).MarkRead(): %s", err)
	}
	rl := len(c1.SearchTitles("asdf", "read"))
	if rl != 1 {
		t.Errorf("category.Search(asdf,read) len: 1 <=> %v", rl)
	}
	print("\t\tMarked\n")
	if err := e1.Mark(); err != nil {
		t.Errorf("entry(1).Mark(): %s", err)
	}
	ml := len(c1.SearchTitles("asdf", "marked"))
	if ml != 1 {
		t.Errorf("category.Search(asdf,marked) len 1 <=> %v", ml)
	}
}
func TestCategory_MarkedEntries(t *testing.T) {
	print("\tCategory.MarkedEntries()\n")
	seed()
	c1 := GetCat("1")
	print("\t\tEmpty\n")
	if el := len(c1.MarkedEntries()); el != 0 {
		t.Errorf("category.MarkedEntries() len 0 <=> %v", el)
	}
	print("\t\tNon-empty\n")
	e1 := GetEntry("1", "test")
	if err := e1.Mark(); err != nil {
		t.Errorf("entry(1).Mark(): %s", err)
	}
	if ml := len(c1.MarkedEntries()); ml != 1 {
		t.Errorf("category(1).MarkedEntries() len 1 <=> %v", ml)
	}
}
func TestCategory_UnreadEntries(t *testing.T) {
	print("\tCategory.UnreadEntries()\n")
	seed()
	c1 := GetCat("1")
	print("\t\tInitial\n")
	if il := len(c1.UnreadEntries()); il != 4 {
		t.Errorf("category(1).UnreadEntries() len 4 <=> %v", il)
	}
	print("\t\tAfter change\n")
	e1 := GetEntry("1", "test")
	if err := e1.MarkRead(); err != nil {
		t.Errorf("entry(1).MarkRead(): %s", err)
	}
	if cl := len(c1.UnreadEntries()); cl != 3 {
		t.Errorf("category(1).UnreadEntries() len 3 <=> %v", cl)
	}

}
func TestCategory_ReadEntries(t *testing.T) {
	print("\tCategory.ReadEntries()\n")
	seed()
	c1 := GetCat("1")
	print("\t\tInitial\n")
	if il := len(c1.ReadEntries()); il != 0 {
		t.Errorf("category(1).ReadEntries() len 0 <=> %v", il)
	}
	print("\t\tAfter change\n")
	e1 := GetEntry("1", "test")
	if err := e1.MarkRead(); err != nil {
		t.Errorf("Error marking read: %s", err)
	}
	if rl := len(c1.ReadEntries()); rl != 1 {
		t.Errorf("category(1).ReadEntries() len 1 <=> %v", rl)
	}
}
func TestCategory_AllEntries(t *testing.T) {
	print("\tCategory.AllEntries()\n")
	seed()
	c1 := GetCat("1")
	if el := len(c1.AllEntries()); el != 4 {
		t.Errorf("category(1).AllEntries() len 4 <=> %v", el)
	}

}
func TestCategory_GetEntriesByParam(t *testing.T) {
	print("\tCategory.GetEntriesByParam()\n")
	seed()
	c1 := GetCat("1")
	if el := len(c1.GetEntriesByParam("1=1")); el != 4 {
		t.Errorf("category(1).GetEntriesByParam(1=1) len 4 <=> %v", el)
	}
}
func TestGetCat(t *testing.T) {
	print("\tCategory.GetCat()\n")
	seed()
	c1 := GetCat("1")
	if c1.ID != 1 {
		t.Errorf("GetCat(1).ID(%v) != 1)", c1.ID)
	}
}
func TestCategory_Feeds(t *testing.T) {
	print("\tCategory.Feeds()\n")
	seed()
	print("\t\tInitial\n")
	c1 := GetCat("1")
	lf := len(c1.Feeds())
	if lf != 2 {
		t.Errorf("c1.Feeds() len 2 <=> %v", lf)
	}
	print("\t\tAdding\n")
	f4, err := GetFeed(4)
	if err != nil {
		t.Errorf("GetFeed: %s", err)
	}
	f4.CategoryID = 1
	f4.Save()
	ls := len(c1.Feeds())
	if ls != 3 {
		t.Errorf("c1.Feeds() len 3 <=> %v", ls)
	}
}
func TestCategory_FeedsStr(t *testing.T) {
	print("\tCategory.FeedsStr()\n")
	seed()
	print("\t\tInitial\n")
	c1 := GetCat("1")
	lf := len(c1.FeedsStr())
	if lf != 2 {
		t.Errorf("c1.FeedsStr() len 2 <=> %v", lf)
	}
	print("\t\tAdding\n")
	f4, err := GetFeed(4)
	if err != nil {
		t.Errorf("GetFeed: %s", err)
	}
	f4.CategoryID = 1
	f4.Save()
	lf = len(c1.FeedsStr())
	if lf != 3 {
		t.Errorf("c1.FeedsStr() len 3 <=> %v", lf)
	}
}
func TestGetCategories(t *testing.T) {
	print("\tCategory.GetCategories(username)\n")
	seed()
	print("\t\tInitial\n")
	il := len(GetCategories("test"))
	if il != 4 {
		t.Errorf("GetCategories(test) len 4 <=> %v", il)
	}
	var newcat Category
	newcat.Name = "newest!"
	newcat.UserName = "test"
	newcat.Insert("test")

	print("\t\tAfterAddition\n")
	nl := len(GetCategories("test"))
	if nl != 5 {
		t.Errorf("GetCategories(test) len 5 <=> %v", nl)
	}
}
func TestCategory_MarkEntriesRead(t *testing.T) {
	print("\tCategory.MarkEntriesRead()\n")
	seed()
	c1 := GetCat("1")
	el := []string{"1", "2"}
	err := c1.MarkEntriesRead(el)
	if err != nil {
		t.Errorf("c1.markEntriesRead([1,2]): %s", err)
	}
	lue := len(c1.UnreadEntries())
	if lue != 2 {
		t.Errorf("c1.UnreadEntries() len 2 <=> %v", lue)
	}
}
func TestCacheAllCats(t *testing.T) {
	print("\tCacheAllCats\n")
	seed()
	var c Category
	var cc int
	print("\t\tInitial\n")
	err := mc.Get("Category1_", &c)
	if err == nil {
		t.Errorf("mc.Get(Category1_) did not error")
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err == nil {
		t.Errorf("mc.Get(Category1_UnreadCount) did not error")
	}
	print("\t\tAfterCaching\n")
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
	print("\t\tInitial\n")
	err := mc.Get("Category1_", &c)
	if err == nil {
		t.Errorf("mc.Get(Category1_) did not error")
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err == nil {
		t.Errorf("mc.Get(Category1_UnreadCount) did not error")
	}
	print("\t\tAfterCaching\n")
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
	print("\t\tAfter Clearing Cache\n")
	err = mc.Get("Category1_", &c)
	if err == nil {
		t.Errorf("mc.Get(Category1_) did not error")
	}
	err = mc.Get("Category1_UnreadCount", &cc)
	if err == nil {
		t.Errorf("mc.Get(Category1_UnreadCount) did not error")
	}

}
func TestGetAllCategories(t *testing.T) {
	print("\tGetAllCategories\n")
	seed()
	ael := len(GetAllCategories())
	if ael != 8 {
		t.Errorf("len(getallCategories()) 8 <=> %v", ael)
	}
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