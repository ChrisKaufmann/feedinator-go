package auth

import (
	"../feed"
	"database/sql"
	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	"github.com/golang/glog"
	"testing"
)

var (
	tmc    = easymemcache.New("127.0.0.1:11211")
	del_us *sql.Stmt
	pop_us *sql.Stmt
	del_ss *sql.Stmt
	pop_ss *sql.Stmt
	del_cc *sql.Stmt
	pop_cc *sql.Stmt
	del_ff *sql.Stmt
	pop_ff *sql.Stmt
)

func init() {
	var err error
	db, err = sql.Open("mysql", "feedinator_test:feedinator_test@tcp(localhost:3306)/feedinator_test")
	if err != nil {
		glog.Fatalf("error: %s", err)
	}
	initDB()
	DB(db)
	mc := &tmc
	del_us, err = u.Sth(db, "delete from users where 1")
	if err != nil {
		glog.Fatalf("delete from users: %s", err)
	}
	pop_us, err = u.Sth(db, "insert into users (id, email) values (1,'user1'),(2,'user2'),(3,'user3'),(4,'user4')")
	if err != nil {
		glog.Fatalf("insert into users: %s", err)
	}
	del_ss, err = u.Sth(db, "delete from sessions where 1")
	if err != nil {
		glog.Fatalf("delete from sessions: %s", err)
	}
	pop_ss, err = u.Sth(db, "insert into sessions (user_id, session_hash) values (1,'user1'),(2,'user2')")
	if err != nil {
		glog.Fatalf("insert into sessions: %s", err)
	}
	if del_cc, err = u.Sth(db, "delete from ttrss_categories"); err != nil {
		glog.Fatalf("delete from categories: %s", err)
	}
	if pop_cc, err = u.Sth(db, "insert into ttrss_categories(id, name, user_name) values (1,'cat1','user1'),(2,'cat2','user1'),(3,'cat3','user1'),(4,'cat4','user1'),(5,'cat5','other'),(6,'cat6','other'),(7,'cat7','other'),(8,'cat8','other')"); err != nil {
		glog.Fatalf("seed categories: %s", err)
	}
	if del_ff, err = u.Sth(db, "delete from ttrss_feeds"); err != nil {
		glog.Fatalf("delete from feeds: %s", err)
	}
	if pop_ff, err = u.Sth(db, "insert into ttrss_feeds (id,title,feed_url,user_name,category_id) values (1,'test1','http://blah','user1',1),(2,'test2','http://blah','user1',1),(3,'test3','http://blah','other',2),(4,'test4','http://blah','other',NULL),(5,'feed','http://blah','user1',NULL),(6,'feed2','http://blah','user1',NULL),(7,'feed3','http://blah','other',NULL),(8,'feed4','http://blah','other',NULL)"); err != nil {
		glog.Fatalf("Seed feeds: %s", err)
	}
	feed.Categoryinit(db, mc)
	feed.Feedinit()

}
func seed() {
	if _, err := del_us.Exec(); err != nil {
		glog.Fatalf("del_us: %s", err)
	}
	if _, err := pop_us.Exec(); err != nil {
		glog.Fatalf("pop_us: %s", err)
	}
	if _, err := del_ss.Exec(); err != nil {
		glog.Fatalf("del_ss: %s", err)
	}
	if _, err := pop_ss.Exec(); err != nil {
		glog.Fatalf("pop_ss: %s", err)
	}
	if _, err := del_cc.Exec(); err != nil {
		glog.Fatalf("del_cc: %s", err)
	}
	if _, err := pop_cc.Exec(); err != nil {
		glog.Fatalf("pop_cc: %s", err)
	}
	if _, err := del_ff.Exec(); err != nil {
		glog.Fatalf("del_ff: %s", err)
	}
	if _, err := pop_ff.Exec(); err != nil {
		glog.Fatalf("pop_ff: %s", err)
	}
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
	if _, err := db.Query("drop table if exists users"); err != nil {
		glog.Fatalf("drop table users: %s", err)
	}
	if _, err := db.Query("CREATE TABLE `users` (id varchar(128), email varchar(128))"); err != nil {
		glog.Fatalf("create table users: %s", err)
	}
	if _, err := db.Query("drop table if exists sessions"); err != nil {
		glog.Fatalf("drop table sessions: %s", err)
	}
	if _, err := db.Query("CREATE TABLE `sessions` (user_id varchar(128), session_hash char(255))"); err != nil {
		glog.Fatalf("create table sessions: %s", err)
	}

}

func TestGetAllUsers(t *testing.T) {
	print("\tGetAllUsers\n")
	seed()
	ul, err := GetAllUsers()
	if err != nil {
		t.Errorf("GetallUsers(): %s", err)
	}
	if len(ul) != 4 {
		t.Errorf("len(GetAllUsers()): 4 <=> %v", len(ul))
	}
}
func TestAddUser(t *testing.T) {
	print("\tAddUser\n")
	seed()
	nul, err := GetAllUsers()
	if err != nil {
		t.Errorf("GetAllUsers(): %s", err)
	}
	ul := len(nul)
	_, err = AddUser("Newest u")
	if err != nil {
		t.Errorf("AddUser(Newest u): %s", err)
	}
	nul, err = GetAllUsers()
	if err != nil {
		t.Errorf("GetAllUsers(): %s", err)
	}
	nl := len(nul)
	if nl != ul+1 {
		t.Errorf("len(GetAllUsers) didn't increase")
	}
}
func TestGetUserByEmail(t *testing.T) {
	print("\tGetUserByEmail\n")
	seed()
	u, err := GetUserByEmail("user1")
	if err != nil {
		t.Errorf("GetUserByEmail(user1): %s", err)
	}
	if u.ID != "1" {
		t.Errorf("user.ID 1 <=> %s", u.ID)
	}
	_, err = GetUserByEmail("nosuchemail")
	if err.Error() != "User Doesn't exist" {
		t.Errorf("Bad error for non-existent user: %s", err)
	}
}
func TestGetUserBySession(t *testing.T) {
	print("\tGetUserBySession\n")
	seed()
	u, err := GetUserBySession("user1")
	if err != nil {
		t.Errorf("GetUserBySession(user1): %s", err)
	}
	if u.ID != "1" {
		t.Error("Bad UserID for returned user by session: %s", u.ID)
	}
	_, err = GetUserBySession("Nosuchsession")
	if err.Error() != "No valid session" {
		t.Errorf("Bad error for no such session: %s", err)
	}
}
func TestSessionExists(t *testing.T) {
	print("\tSessionExists\n")
	seed()
	if !SessionExists("user1") {
		t.Error("SessionExists(user1) returned false")
	}
	if SessionExists("nosuchsession") {
		t.Error("SessionExists(nosuchsession) returned true")
	}
}
func TestUserExists(t *testing.T) {
	print("\tUserExists\n")
	seed()
	if !UserExists("user1") {
		t.Error("userExists(user1) returned false")
	}
	if UserExists("nouser") {
		t.Error("userExists(nouser) returned true")
	}
}

func TestUser_AddSession(t *testing.T) {
	print("\tUser.AddSession()\n")
	seed()
	u, err := GetUserByEmail("user1")
	if err != nil {
		t.Errorf("GetUserByEmail(user1): %s", err)
	}
	err = u.AddSession("newsession")
	if err != nil {
		t.Errorf("u.AddSession(newsession): %s", err)
	}

	u2, err := GetUserBySession("newsession")
	if err != nil {
		t.Errorf("GetUserBySession(newsession): %s", err)
	}
	if u2.ID != "1" {
		t.Error("GetUserBySession(newsession).ID != 1: %s", u2.ID)
	}

}
func TestUser_Categories(t *testing.T) {
	print("\tUser.Categories()\n")
	seed()
	u, err := GetUserByEmail("user1")
	if err != nil {
		t.Errorf("GetUserByEmail(user1): %s", err)
	}
	cl := len(u.Categories())
	if cl != 4 {
		t.Errorf("len(user1.Categories()): 4 <=> %v", cl)
	}
}
func TestUser_Feeds(t *testing.T) {
	print("\tUser.Feeds()\n")
	seed()
	u, err := GetUserByEmail("user1")
	if err != nil {
		t.Errorf("GetUserByEmail(user1): %s", err)
	}
	fl := len(u.Feeds())
	if fl != 4 {
		t.Errorf("len(user1.Feeds()): 4 <=> %v", fl)
	}

}
