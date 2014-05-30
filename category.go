package main

import (
	"strings"
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
	sid := strconv.Itoa(c.ID)
	cl := []string{"Category"+sid,"CategoryUnreadCount"+sid,"CategoryFeeds_" + sid,"CategoryList_" + userName,"CategoryList"  }
	for _,i := range cl {
		err := mc.Delete(i)
		if err != nil {
			err.Error()
		}
	}
}
func (c Category) Unread() int {
	var count int
	var cct = "CategoryUnreadCount" + strconv.Itoa(c.ID)
	unreadc, err := mc.Get(cct)
	if err != nil {
		var query = "select count(*) from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and unread='1'"
		var stmt = sth(db,query)
		err := stmt.QueryRow().Scan(&count)
		if err != nil {
			err.Error()
		}
		mcsettime(cct, count, 60)
	} else {
		err = json.Unmarshal(unreadc.Value, &count)
		print("+cuc" + strconv.Itoa(c.ID))
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
	feeds := c.Feeds()
	var feedstr []string
	var el []Entry
	fl := make(map[string]Feed)
    for _,f := range feeds {
        feedstr = append(feedstr, strconv.Itoa(f.ID))
		fl[strconv.Itoa(f.ID)]=f
    }
	var query = "select e.id,IFNULL(e.title,''),IFNULL(e.link,''),IFNULL(e.updated,''),e.marked,e.unread,e.feed_id from ttrss_entries e where e.feed_id in (" + strings.Join(feedstr, ", ") + ") and " + p + " order by e.id ASC;"
	var stmt = sth(db, query)
	rows, err := stmt.Query()
	if err != nil {
		err.Error()
	}
	var count int
	for rows.Next() {
		var e Entry
		rows.Scan(&e.ID, &e.Title, &e.Link, &e.Date, &e.Marked, &e.Unread, &e.FeedID)
		e.Evenodd = evenodd(count)
		f := fl[strconv.Itoa(e.FeedID)]
		e.FeedName = f.Title
		e = e.Normalize()
		el = append(el, e)
		count = count + 1
	}
	return el
}
func (c Category) Next (id string) Entry {
	var retval string
	nes := "CategoryEntry"+id+"Next"
	ne, err := mc.Get(nes)
	var e Entry
	if err != nil {
		print("-"+nes)
		var query = "select id from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and id > "+id+" order by id ASC limit 1"
		var stmt = sth(db,query)
		stmt.QueryRow().Scan(&retval)
		e = getEntry(retval)
		mcsettime(nes, e, 300)
	} else {
		print("+"+nes)
		err = json.Unmarshal(ne.Value, &e)
	}
	return e
}
func (c Category) Previous(id string) Entry {
	var e Entry
	pes := "CategoryEntry"+id+"Previous"
	pe, err := mc.Get(pes)
	if err != nil {
		print("-"+pes)
		var retval string
		var query = "select id from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and id < "+id+" order by id DESC limit 1"
		var stmt = sth(db,query)
		stmt.QueryRow().Scan(&retval)
		e := getEntry(retval)
		mcsettime(pes, e, 300)
		return e
	} else {
		print("+"+pes)
		err = json.Unmarshal(pe.Value, &e)
	}
	return e
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
		for _,i := range feedids {
			feed := getFeed(strconv.Itoa(i))
			allFeeds = append(allFeeds, feed)
		}
	}
	return allFeeds
}
func (c Category) FeedsStr() []string {
	f := c.Feeds()
	var feedstr []string
	for _,i := range f {
		feedstr = append(feedstr, strconv.Itoa(i.ID))
	}
	return feedstr
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
		for _,i := range catids {
			cat := getCat(strconv.Itoa(i))
			allCats = append(allCats, cat)
		}

	}
	return allCats
}
func GetAllCategories() []Category {
	var allCats []Category
	var catids []int
	catlist, err := mc.Get("CategoryList")
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
			mcset("Category"+strconv.Itoa(cat.ID), cat)
		}
		mcset("CategoryList", allCats)
	} else {
		print("+CL<ALL>")
		err = json.Unmarshal(catlist.Value, allCats)
		for _,i := range catids {
			cat := getCat(strconv.Itoa(i))
			allCats = append(allCats, cat)
		}
	}
	return allCats
}
func cacheAllCats() {
	_ = GetAllCategories()
}
