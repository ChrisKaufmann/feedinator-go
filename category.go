package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html"
	"strings"
)

type Category struct {
	Name        string
	Description string
	UserName    string
	ID          int
	Evenodd     string
	Exclude     string
}

var (
	stmtGetCat          *sql.Stmt
	stmtGetCats         *sql.Stmt
	stmtGetAllCats      *sql.Stmt
	stmtGetCatFeeds     *sql.Stmt
	stmtSaveCat         *sql.Stmt
	stmtAddCat          *sql.Stmt
	stmtResetCategories *sql.Stmt
	stmtDeleteCategory  *sql.Stmt
)

func init() {
	stmtGetCatFeeds = sth(db, "select id from ttrss_feeds where category_id = ?")
	stmtGetCat = sth(db, "select name,user_name,IFNULL(description,''),id, exclude from ttrss_categories where id = ?")
	stmtGetCats = sth(db, "select name,user_name,IFNULL(description,''),id,exclude from ttrss_categories where user_name= ?")
	stmtGetAllCats = sth(db, "select name,user_name,IFNULL(description,''),id,exclude from ttrss_categories")
	stmtSaveCat = sth(db, "update ttrss_categories set name=?,description=?, exclude=? where id=? limit 1")
	stmtAddCat = sth(db, "insert into ttrss_categories (user_name,name) values (?,?)")
	stmtResetCategories = sth(db, "update ttrss_feeds set category_id=NULL where category_id= ?")
	stmtDeleteCategory = sth(db, "delete from ttrss_categories where id=? limit 1")
}
func (c Category) Save() {
	if c.Description == "" {
		c.Description = " "
	}
	c.Exclude = html.EscapeString(c.Exclude)
	_, err := stmtSaveCat.Exec(c.Name, c.Description, c.Exclude, c.ID)
	if err != nil {
		err.Error()
	}
	c.ClearCache()
}
func (c Category) Print() {
	print("Category: " + tostr(c.ID) + "\n" +
		"\tName:\t" + c.Name + "\n" +
		"\tDesc:\t" + c.Description + "\n" +
		"\tUser:\t" + c.UserName + "\n" +
		"\tExclude:\t" + c.Exclude + "\n")
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
	for _, i := range fl {
		i.Update()
	}
	c.ClearCache()
}
func (c Category) ClearCache() {
	mc.DeleteLike("Category" + tostr(c.ID) + "_")
	cl := mc.Find("Category" + tostr(c.ID))
	mc.Delete(cl...)
	mcl := []string{"Category" + tostr(c.ID), "Category" + tostr(c.ID) + "_UnreadCount", "Category" + tostr(c.ID) + "_Feeds"}
	for _, k := range mcl {
		print("Deleting " + k + "\n")
		mc.Delete(k)
	}
	mc.Delete(mcl...)
}
func (c Category) Unread() (count int) {
	mc.GetOr("Category"+tostr(c.ID)+"_UnreadCount", &count, func() {
		print("cc" + tostr(c.ID) + "-")
		if len(c.FeedsStr()) < 1 {
			count = 0
			return
		}
		var query = "select count(*) from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and unread='1'"
		var stmt = sth(db, query)
		err := stmt.QueryRow().Scan(&count)
		if err != nil {
			print(err.Error())
		}
	})
	return count
}
func (c Category) Class() string {
	if c.Unread() > 0 {
		return "oddUnread"
	}
	return "odd"
}
func (c Category) Excludes() []string {
	return strings.Split(strings.ToLower(c.Exclude), ",")
}
func (c Category) DeleteExcludes() {
	for _, f := range c.Feeds() {
		f.DeleteExcludes()
	}
}
func (c Category) SearchTitles(s string) (el []Entry) {
	// start with unread
	//	mc.GetOr("Category"+tostr(c.ID)+"_search_"+s, &el, func() {
	//		el = c.GetEntriesByParam("title like '%"+s+"%'")
	//	})
	ul := c.UnreadEntries()
	for _, e := range ul {
		if strings.Contains(strings.ToLower(e.Title), strings.ToLower(s)) {
			el = append(el, e)
		}
	}

	return el
}
func (c Category) MarkedEntries() (el []Entry) {
	mc.GetOr("Category"+tostr(c.ID)+"_markedentries", &el, func() {
		el = c.GetEntriesByParam("marked = 1")
	})
	return el
}
func (c Category) UnreadEntries() (el []Entry) {
	mc.GetOr("Category"+tostr(c.ID)+"_unreadentries", &el, func() {
		el = c.GetEntriesByParam("unread = 1")
	})
	mc.Set("Category"+tostr(c.ID)+"_UnreadCount", len(el))
	return el
}
func (c Category) ReadEntries() (el []Entry) {
	mc.GetOr("Category"+tostr(c.ID)+"_readentries", &el, func() {
		print(".")
		el = c.GetEntriesByParam("unread = '0'")
	})
	return el
}
func (c Category) AllEntries() (el []Entry) {
	el = c.GetEntriesByParam("1=1")
	return el
}
func (c Category) GetEntriesByParam(p string) (el []Entry) {
	if len(c.FeedsStr()) < 1 {
		return el
	}
	var query = "select " + entrySelectString + " from ttrss_entries  where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and " + p + " order by id ASC;"
	el = getEntriesFromSql(query)
	return el
}
func (c Category) Next(id string) Entry {
	var retval string
	var e Entry
	nes := "Category" + tostr(c.ID) + "NextEntry" + id
	err := mc.Get(nes, &e)
	if err != nil {
		print("-" + nes)
		var query = "select id from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and id > " + id + " order by id ASC limit 1"
		var stmt = sth(db, query)
		stmt.QueryRow().Scan(&retval)
		e = getEntry(retval)
		mc.SetTime(nes, e, 300)
	}
	return e
}
func (c Category) Previous(id string) Entry {
	var e Entry
	pes := "CategoryEntry" + tostr(c.ID) + "PreviousEntry" + id
	err := mc.Get(pes, &e)
	if err != nil {
		print("-" + pes)
		var retval string
		var query = "select id from ttrss_entries where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and id < " + id + " order by id DESC limit 1"
		var stmt = sth(db, query)
		stmt.QueryRow().Scan(&retval)
		e := getEntry(retval)
		mc.SetTime(pes, e, 300)
		return e
	}
	return e
}

func getCat(id string) Category {
	var cat Category
	err := mc.Get("Category"+id, &cat)
	if err != nil { //cache miss
		err := stmtGetCat.QueryRow(id).Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID, &cat.Exclude)
		if err != nil {
			err.Error()
		}
		print("-cat" + tostr(cat.ID))
		mc.Set("Category"+id, cat)
	}
	return cat
}
func (c Category) Feeds() []Feed {
	var allFeeds []Feed
	var feedids []int
	var cfl = "Category" + tostr(c.ID) + "Feeds"

	//Try getting from cache first
	err := mc.Get(cfl, &feedids)
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
		mc.Set(cfl, feedids)
	} else {
		print("+CFL_" + tostr(c.ID))
		for _, i := range feedids {
			feed := getFeed(tostr(i))
			allFeeds = append(allFeeds, feed)
		}
	}
	return allFeeds
}
func (c Category) FeedsStr() []string {
	f := c.Feeds()
	var feedstr []string
	for _, i := range f {
		feedstr = append(feedstr, tostr(i.ID))
	}
	return feedstr
}

func getCategories() []Category {
	var allCats []Category
	var catids []int

	//Try getting a category list from cache
	err := mc.Get("CategoryList_"+userName, &catids)
	if err != nil {
		print("-CL" + userName)
		rows, err := stmtGetCats.Query(userName)
		if err != nil {
			err.Error()
			return allCats
		}
		for rows.Next() {
			var cat Category
			rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID, &cat.Exclude)
			allCats = append(allCats, cat)
			catids = append(catids, cat.ID)
			mc.Set("Category"+tostr(cat.ID), cat)
		}
		mc.Set("CategoryList_"+userName, catids)
	} else {
		print("+CL" + userName)
		for _, i := range catids {
			cat := getCat(tostr(i))
			allCats = append(allCats, cat)
		}

	}
	return allCats
}
func GetAllCategories() []Category {
	var allCats []Category
	var catids []int
	err := mc.Get("CategoryList", &catids)
	if err != nil {
		print("-CL<ALL>")
		rows, err := stmtGetAllCats.Query()
		if err != nil {
			err.Error()
			return allCats
		}
		for rows.Next() {
			var cat Category
			rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID, &cat.Exclude)
			allCats = append(allCats, cat)
			catids = append(catids, cat.ID)
			mc.Set("Category"+tostr(cat.ID), cat)
		}
		mc.Set("CategoryList", allCats)
	} else {
		print("+CL<ALL>")
		for _, i := range catids {
			cat := getCat(tostr(i))
			allCats = append(allCats, cat)
		}
	}
	return allCats
}
func cacheAllCats() {
	_ = GetAllCategories()
}
