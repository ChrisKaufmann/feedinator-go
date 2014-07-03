package main

import (
	"strings"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Category struct {
	Name        string
	Description string
	UserName    string
	ID          int
	Evenodd     string
}

var (
	stmtGetCat                *sql.Stmt
	stmtGetCats               *sql.Stmt
	stmtGetAllCats            *sql.Stmt
	stmtGetCatFeeds           *sql.Stmt
	stmtSaveCat               *sql.Stmt
	stmtAddCat                *sql.Stmt
	stmtResetCategories       *sql.Stmt
	stmtDeleteCategory        *sql.Stmt
)

func init() {
	stmtGetCatFeeds = sth(db, "select id from ttrss_feeds where category_id = ?")
	stmtGetCat = sth(db, "select name,user_name,IFNULL(description,''),id from ttrss_categories where id = ?")
	stmtGetCats = sth(db, "select name,user_name,IFNULL(description,''),id from ttrss_categories where user_name= ?")
	stmtGetAllCats = sth(db, "select name,user_name,IFNULL(description,''),id from ttrss_categories")
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
	for _,i := range fl {
		i.Update()
	}
	c.ClearCache()
}
func (c Category) ClearCache() {
	sid := tostr(c.ID)
	cl := []string{"Category"+sid,"CategoryUnreadCount"+sid,"CategoryFeeds_" + sid,"CategoryList_" + userName,"CategoryList"  }
	for _,i := range cl {
//		err := mc.Delete(i)
		err := mcdel(i)
		if err != nil {
			err.Error()
		}
	}
}
func (c Category) Unread() int {
	var count int
	var cct = "CategoryUnreadCount" + tostr(c.ID)
	err := mcget(cct, &count)
	if err != nil {
		var query = "select count(*) from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and unread='1'"
		var stmt = sth(db,query)
		err := stmt.QueryRow().Scan(&count)
		if err != nil {
			err.Error()
		}
		mcsettime(cct, count, 60)
	}
	return count
}
func (c Category) Class() string {
	if c.Unread() > 0 {
		return "oddUnread"
	}
	return "odd"
}
func (c Category) MarkedEntries() []Entry {
	el := c.GetEntriesByParam("marked = 1")
	return el
}
func (c Category) UnreadEntries() []Entry {
	el := c.GetEntriesByParam("unread = 1")
	return el
}
func (c Category) ReadEntries() []Entry {
	el := c.GetEntriesByParam("unread = '0'")
	return el
}
func (c Category) AllEntries() []Entry {
	el := c.GetEntriesByParam("1=1")
	return el
}
func (c Category) GetEntriesByParam(p string) []Entry {
	var query = "select "+entrySelectString +" from ttrss_entries  where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and " + p + " order by id ASC;"
	el := getEntriesFromSql(query)
	return el
}
func (c Category) Next (id string) Entry {
	var retval string
	var e Entry
	nes := "CategoryEntry"+id+"Next"
	err := mcget(nes, &e)
	if err != nil {
		print("-"+nes)
		var query = "select id from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and id > "+id+" order by id ASC limit 1"
		var stmt = sth(db,query)
		stmt.QueryRow().Scan(&retval)
		e = getEntry(retval)
		mcsettime(nes, e, 300)
	}
	return e
}
func (c Category) Previous(id string) Entry {
	var e Entry
	pes := "CategoryEntry"+id+"Previous"
	err := mcget(pes, &e)
	if err != nil {
		print("-"+pes)
		var retval string
		var query = "select id from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and id < "+id+" order by id DESC limit 1"
		var stmt = sth(db,query)
		stmt.QueryRow().Scan(&retval)
		e := getEntry(retval)
		mcsettime(pes, e, 300)
		return e
	}
	return e
}

func getCat(id string) Category {
	var cat Category
	err := mcget("Category" + id,&cat)
	if err != nil { //cache miss
		err := stmtGetCat.QueryRow(id).Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
		if err != nil {
			err.Error()
		}
		print("-cat" + tostr(cat.ID))
		mcset("Category"+id, cat)
	}
	return cat
}
func (c Category) Feeds() []Feed {
	var allFeeds []Feed
	var feedids []int

	//Try getting from cache first
	err := mcget("CategoryFeeds_" + tostr(c.ID), &feedids)
	if err != nil {
		rows, err := stmtGetCatFeeds.Query(tostr(c.ID))
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
		mcset("CategoryFeeds_"+tostr(c.ID), feedids)
	} else {
		print("+CFL_" + tostr(c.ID))
		for _,i := range feedids {
			feed := getFeed(tostr(i))
			allFeeds = append(allFeeds, feed)
		}
	}
	return allFeeds
}
func (c Category) FeedsStr() []string {
	f := c.Feeds()
	var feedstr []string
	for _,i := range f {
		feedstr = append(feedstr, tostr(i.ID))
	}
	return feedstr
}

func getCategories() []Category {
	var allCats []Category
	var catids []int

	//Try getting a category list from cache
	err := mcget("CategoryList_" + userName, &catids)
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
			mcset("Category"+tostr(cat.ID), cat)
		}
		mcset("CategoryList_"+userName, catids)
	} else {
		print("+CL" + userName)
		for _,i := range catids {
			cat := getCat(tostr(i))
			allCats = append(allCats, cat)
		}

	}
	return allCats
}
func GetAllCategories() []Category {
	var allCats []Category
	var catids []int
	err := mcget("CategoryList", &catids)
	if err != nil {
		print("-CL<ALL>")
		rows, err := stmtGetAllCats.Query()
		if err != nil {
			err.Error()
			return allCats
		}
		for rows.Next() {
			var cat Category
			rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
			allCats = append(allCats, cat)
			catids = append(catids, cat.ID)
			mcset("Category"+tostr(cat.ID), cat)
		}
		mcset("CategoryList", allCats)
	} else {
		print("+CL<ALL>")
		for _,i := range catids {
			cat := getCat(tostr(i))
			allCats = append(allCats, cat)
		}
	}
	return allCats
}
func cacheAllCats() {
	_ = GetAllCategories()
}
