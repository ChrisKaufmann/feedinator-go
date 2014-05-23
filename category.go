package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

type Category struct {
	Name        string
	Description string
	UserName    string
	ID          int
	Evenodd     string
}

var (
	stmtCatList               *sql.Stmt
	stmtCatUnread             *sql.Stmt
	stmtGetCat                *sql.Stmt
	stmtGetCats               *sql.Stmt
	stmtNextCategoryEntry     *sql.Stmt
	stmtPreviousCategoryEntry *sql.Stmt
	stmtGetCatFeeds           *sql.Stmt
	stmtGetFeedsInCat         *sql.Stmt
	stmtSaveCat               *sql.Stmt
	stmtAddCat                *sql.Stmt
	stmtResetCategories       *sql.Stmt
	stmtDeleteCategory        *sql.Stmt
)

func init() {
	stmtCatList = sth(db, "select name,id from ttrss_categories where user_name=?")
	stmtCatUnread = sth(db, "select count(e.id) as unread from ttrss_entries as e,ttrss_feeds as f where f.category_id= ? and e.feed_id=f.id and e.unread='1'")
	stmtGetCatFeeds = sth(db, "select f.id from ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and c.id= ?")
	stmtGetCat = sth(db, "select name,user_name,IFNULL(description,''),id from ttrss_categories where id = ?")
	stmtGetCats = sth(db, "select name,user_name,IFNULL(description,''),id from ttrss_categories where user_name= ?")
	stmtNextCategoryEntry = sth(db, "select e.id from ttrss_entries as e,ttrss_feeds as f  where f.category_id=? and e.feed_id=f.id and e.id > ? order by e.id ASC limit 1")
	stmtPreviousCategoryEntry = sth(db, "select e.id from ttrss_entries as e, ttrss_feeds as f where f.category_id=? and e.feed_id=f.id and e.id<? order by e.id DESC limit 1")
	stmtSaveCat = sth(db, "update ttrss_categories set name=?,description=? where id=? limit 1")
	stmtAddCat = sth(db, "insert into ttrss_categories (user_name,name) values (?,?)")
	stmtResetCategories = sth(db, "update ttrss_feeds set category_id=NULL where category_id= ?")
	stmtDeleteCategory = sth(db, "delete from ttrss_categories where id=? limit 1")
}
func (c Category) Save() {
	if c.Description == "" {
		c.Description = " "
	}
	stmtSaveCat.Exec(c.Name, c.Description, c.ID)
	c.ClearCache()
}
func (c Category) Insert() {
	stmtAddCat.Exec(userName, c.Name)
}
func (c Category) Delete() {
	stmtResetCategories.Exec(c.ID)
	stmtDeleteCategory.Exec(c.ID)
	c.ClearCache()
}
func (c Category) Update() {
	fl := c.Feeds()
	for i := range fl {
		fl[i].Update()
	}
	c.ClearCache()
}
func (c Category) ClearCache() {
	err := mc.Delete("Category"+strconv.Itoa(c.ID))
	if err != nil {
		err.Error()
	}
	err = mc.Delete("CategoryUnreadCount"+strconv.Itoa(c.ID))
	if err != nil {
		err.Error()
	}
}
func (c Category) Unread() int {
	var count int
	var cct = "CategoryUnreadCount"+strconv.Itoa(c.ID)
	unreadc, err := mc.Get(cct)
	if err != nil {
		err := stmtCatUnread.QueryRow(c.ID).Scan(&count)
		if err != nil {
			err.Error()
		}
		mcsettime(cct,count,60)
	} else {
		err = json.Unmarshal(unreadc.Value, &count)
		print("+cuc"+strconv.Itoa(c.ID))
	}
	return count
}
func (c Category) Class() string {
	if c.Unread() > 0 {
		return "oddUnread"
	}
	return "odd"
}

func getCat(id string) Category {
	var cat Category
	nucat, err := mc.Get("Category" + id)
	if err != nil { //cache miss
		err := stmtGetCat.QueryRow(id).Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
		if err != nil {
			err.Error()
		}
		print("-cat" + strconv.Itoa(cat.ID))
		mcset("Category"+id, cat)
	} else {
		err = json.Unmarshal(nucat.Value, &cat)
		print("+cat" + strconv.Itoa(cat.ID))
	}
	return cat
}
func (c Category) Feeds() []Feed {
	var allFeeds []Feed
	var feedids []int

	//Try getting from cache first
	catfeeds, err := mc.Get("CategoryFeeds_" + strconv.Itoa(c.ID))
	if err != nil {
		rows, err := stmtGetCatFeeds.Query(strconv.Itoa(c.ID))
		if err != nil {
			err.Error()
			return allFeeds
		}
		for rows.Next() {
			var id string
			rows.Scan(&id)
			feed := getFeed(id)
			allFeeds = append(allFeeds, feed)
			feedids = append(feedids, feed.ID)
		}
		mcset("CategoryFeeds_"+strconv.Itoa(c.ID), feedids)
	} else {
		print("+CFL_" + strconv.Itoa(c.ID))
		err = json.Unmarshal(catfeeds.Value, &feedids)
		for i := range feedids {
			feed := getFeed(strconv.Itoa(feedids[i]))
			allFeeds = append(allFeeds, feed)
		}
	}
	return allFeeds
}

func getCategories() []Category {
	var allCats []Category
	var catids []int

	//Try getting a category list from cache
	catlist, err := mc.Get("CategoryList_" + userName)
	if err != nil {
		print("-CL" + userName)
		rows, err := stmtGetCats.Query(userName)
		if err != nil {
			err.Error()
			return allCats
		}
		for rows.Next() {
			var cat Category
			rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
			allCats = append(allCats, cat)
			catids = append(catids, cat.ID)
			mcset("Category"+strconv.Itoa(cat.ID), cat)
		}
		mcset("CategoryList_"+userName, catids)
	} else {
		print("+CL" + userName)
		err = json.Unmarshal(catlist.Value, &catids)
		for i := range catids {
			cat := getCat(strconv.Itoa(catids[i]))
			allCats = append(allCats, cat)
		}

	}
	return allCats
}
