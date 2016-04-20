package feed

import (
	"database/sql"
	"fmt"
	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"html"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Category struct {
	Name         string
	Description  string
	UserName     string
	ID           int
	Evenodd      string
	Exclude      string
	SearchSelect template.HTML
	Search       string
}

var (
	stmtGetCat          *sql.Stmt
	stmtGetCats         *sql.Stmt
	stmtGetAllCats      *sql.Stmt
	stmtSaveCat         *sql.Stmt
	stmtAddCat          *sql.Stmt
	stmtResetCategories *sql.Stmt
	stmtDeleteCategory  *sql.Stmt
	db                  *sql.DB
	mc                  *easymemcache.Client
)

func (c Category) String() string {
	return fmt.Sprintf("ID: %s,Name: %s, Description: %s, UserName: %s, Exclude: %s", c.ID, c.Name, c.Description, c.UserName, c.Exclude)
}

func Categoryinit(dbh *sql.DB, mch *easymemcache.Client) {
	var err error
	db = dbh
	mc = mch
	stmtGetCat, err = u.Sth(db, "select name,user_name,IFNULL(description,''),id, exclude from ttrss_categories where id = ?")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
	stmtGetCats, err = u.Sth(db, "select name,user_name,IFNULL(description,''),id,exclude from ttrss_categories where user_name= ?")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
	stmtGetAllCats, err = u.Sth(db, "select name,user_name,IFNULL(description,''),id,exclude from ttrss_categories")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
	stmtSaveCat, err = u.Sth(db, "update ttrss_categories set name=?,description=?, exclude=? where id=? limit 1")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
	stmtAddCat, err = u.Sth(db, "insert into ttrss_categories (user_name,name) values (?,?)")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
	stmtResetCategories, err = u.Sth(db, "update ttrss_feeds set category_id=NULL where category_id= ?")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
	stmtDeleteCategory, err = u.Sth(db, "delete from ttrss_categories where id=? limit 1")
	if err != nil {
		glog.Fatalf("sth(): %s", err)
	}
}
func (c Category) Save() (err error) {
	if c.Description == "" {
		c.Description = " "
	}
	c.Exclude = html.EscapeString(c.Exclude)
	_, err = stmtSaveCat.Exec(c.Name, c.Description, c.Exclude, c.ID)
	if err != nil {
		glog.Errorf("stmtSaveCat.Exec(%s,%s,%s,%s): %s", c.Name, c.Description, c.Exclude, c.ID, err)
	}
	c.ClearCache()
	return err
}
func (c Category) Print() {
	print("Category: " + u.Tostr(c.ID) + "\n" +
		"\tName:\t" + c.Name + "\n" +
		"\tDesc:\t" + c.Description + "\n" +
		"\tUser:\t" + c.UserName + "\n" +
		"\tExclude:\t" + c.Exclude + "\n")
}
func (c Category) Insert(userName string) (err error) {
	_, err = stmtAddCat.Exec(userName, c.Name)
	if err != nil {
		glog.Errorf("stmtAddCat.Exec(%s,%s): %s", userName, c.Name, err)
	}
	mc.Delete("CategoryList_" + userName)
	mc.Delete("CategoryList")
	return err
}
func (c Category) Delete() (err error) {
	_, err = stmtResetCategories.Exec(c.ID)
	if err != nil {
		glog.Errorf("stmtResetCategories.Exec(%s): %s", c.ID, err)
		return err
	}
	_, err = stmtDeleteCategory.Exec(c.ID)
	if err != nil {
		glog.Errorf("stmtDeleteCategory.Exec(%s): %s", c.ID, err)
		return err
	}
	mc.Delete("CategoryList_" + c.UserName)
	mc.Delete("CategoryList")
	c.ClearCache()
	return err
}
func (c Category) Update() {
	fl := c.Feeds()
	for _, i := range fl {
		i.Update()
	}
	c.ClearCache()
}
func (c Category) ClearCache() {
	mc.DeleteLike("Category" + u.Tostr(c.ID) + "_")
	cl := mc.Find("Category" + u.Tostr(c.ID))
	mc.Delete(cl...)
	mcl := []string{"Category" + u.Tostr(c.ID), "Category" + u.Tostr(c.ID) + "_UnreadCount", "Category" + u.Tostr(c.ID) + "_Feeds"}
	for _, k := range mcl {
		mc.Delete(k)
	}
	mc.Delete(mcl...)
}
func (c Category) Unread() (count int) {
	t0 := time.Now()
	mc.GetOr("Category"+u.Tostr(c.ID)+"_UnreadCount", &count, func() {
		for _, f := range c.Feeds() {
			count = count + f.Unread()
		}
	})
	fmt.Printf("Category(%v).Unread(): %s", c.ID, time.Now().Sub(t0))
	return count
}
func (c Category) Class() string {
	if c.Unread() > 0 {
		return "oddUnread"
	}
	return "odd"
}
func (c Category) Excludes() (rl []string) {
	for _, e := range strings.Split(strings.ToLower(c.Exclude), ",") {
		if len(e) > 0 {
			rl = append(rl, e)
		}
	}
	return rl
}
func (c Category) DeleteExcludes() {
	//Go through the included feeds and delete their excludes first
	for _, f := range c.Feeds() {
		//There are weird things with caching of feeds if it has recently been updated.
		fd, err := GetFeed(f.ID)
		if err != nil {
			glog.Errorf("GetFeed(%v): %s", f.ID, err)
			continue
		}
		fd.DeleteExcludes()
	}
	for _, excstr := range c.Excludes() {
		if len(excstr) < 1 {
			continue
		}
		var query = fmt.Sprintf("delete from ttrss_entries where feed_id in (%s) and title like '%%%s%%'", strings.Join(c.FeedsStr(), ","), excstr)
		if _, err := db.Query(query); err != nil {
			glog.Errorf("Category(%v).DeleteExcludes.delete(%s): %s", c.ID, excstr, err)
		}
	}
	c.ClearCache()
}
func (c Category) SearchTitles(s string, m string) (el []Entry) { //s=search string, m=modifier (read/unread/marked/all)
	var ul []Entry
	var ss string
	switch m {
	case "marked":
		ss = fmt.Sprintf(" title like '%%%s%%' and marked = '1'", s)
	case "read":
		ss = fmt.Sprintf(" title like '%%%s%%' and unread='0'", s)
	case "all":
		ss = fmt.Sprintf(" title like '%%%s%%' ", s)
	default: //yeah, default to unread
		ss = fmt.Sprintf(" title like '%%%s%%' and unread=1", s)
	}
	if s == "" {
		return ul
	}
	ul = c.GetEntriesByParam(ss)
	if len(ul) != 0 {
		for _, e := range ul {
			if strings.Contains(strings.ToLower(e.Title), strings.ToLower(s)) {
				el = append(el, e)
			}
		}
	}
	mc.Set("CategoryCurrent"+c.UserName, el)
	return el
}
func (c Category) MarkedEntries() (el []Entry) {
	mc.GetOr("Category"+u.Tostr(c.ID)+"_markedentries", &el, func() {
		t0 := time.Now()
		for _, f := range c.Feeds() {
			el = append(el, f.MarkedEntries()...)
		}
		sort.Sort(EntryList(el))
		fmt.Printf("Search Feeds: %v", time.Now().Sub(t0))
		t0 = time.Now()
		el = c.GetEntriesByParam("marked = 1")
		fmt.Printf("By Category: %v", time.Now().Sub(t0))
	})
	return el
}
func (c Category) UnreadEntries() (el []Entry) {
	mc.GetOr("Category"+u.Tostr(c.ID)+"_unreadentries", &el, func() {
		for _, f := range c.Feeds() {
			el = append(el, f.UnreadEntries()...)
		}
		sort.Sort(EntryList(el))
	})
	mc.SetTime("Category"+u.Tostr(c.ID)+"_UnreadCount", len(el), 60)
	return el
}
func (c Category) ReadEntries() (el []Entry) {
	mc.GetOr("Category"+u.Tostr(c.ID)+"_readentries", &el, func() {
		for _, f := range c.Feeds() {
			el = append(el, f.ReadEntries()...)
		}
		sort.Sort(EntryList(el))
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
	var query = "select " + entrySelectString + " from ttrss_entries  where feed_id in (" + strings.Join(c.FeedsStr(), ", ") + ") and " + p + " order by id ASC limit 500;"
	el = getEntriesFromSql(query)
	mc.Set("CategoryCurrent"+c.UserName, el)
	return el
}
func (c Category) MarkEntriesRead(ids []string) (err error) {
	if len(ids) == 0 {
		err = fmt.Errorf("Ids is null")
	} else {
		// make sure they're all integers
		var id_list []string
		for _, i := range ids {
			_, err := strconv.Atoi(i)
			if err != nil {
				glog.Errorf("Non int passed to markEntriesRead: %s", err)
			} else {
				id_list = append(id_list, i)
			}
		}
		if len(id_list) < 1 {
			err = fmt.Errorf("Not enough valid ids passed")
			return err
		}
		j := strings.Join(id_list, ",")
		sql := "update ttrss_entries set unread=0 where id in (" + j + ")"
		stmtUpdateMarkEntries, err := u.Sth(db, sql)
		if err != nil {
			glog.Errorf("Sth(db, %s): %s", sql, err)
			return err
		}
		_, err = stmtUpdateMarkEntries.Exec()
		if err != nil {
			glog.Errorf("stmtUpdateMarkEntries.Exec: %s", err)
			return err
		}
		mc.Decrement("Category"+u.Tostr(c.ID)+"_UnreadCount", uint64(len(ids)))
		mc.Delete("Category" + u.Tostr(c.ID) + "_unreadentries")
		mc.Delete("Category" + u.Tostr(c.ID) + "_readentries")
		for _, f := range c.Feeds() {
			mc.Delete("Feed" + u.Tostr(f.ID) + "_readentries")
			mc.Delete("Feed" + u.Tostr(f.ID) + "_UnreadCount")
			mc.Delete("Feed" + u.Tostr(f.ID) + "_unreadentries")
		}
	}
	return err
}
func (c Category) Feeds() []Feed {
	af := GetCategoryFeeds(c.ID)
	return af
}
func (c Category) FeedsStr() []string {
	f := c.Feeds()
	var feedstr []string
	for _, i := range f {
		feedstr = append(feedstr, u.Tostr(i.ID))
	}
	return feedstr
}

func GetCat(id string) Category {
	var cat Category
	err := mc.Get("Category"+id+"_", &cat)
	if err != nil { //cache miss
		err := stmtGetCat.QueryRow(id).Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID, &cat.Exclude)
		if err != nil {
			err.Error()
		}
		mc.Set("Category"+id+"_", cat)
	}
	return cat
}
func GetCategories(userName string) []Category {
	t0 := time.Now()
	var allCats []Category
	var catids []int

	//Try getting a category list from cache
	err := mc.Get("CategoryList_"+userName, &catids)
	if err != nil {
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
			mc.Set("Category"+u.Tostr(cat.ID), cat)
		}
		mc.Set("CategoryList_"+userName, catids)
	} else {
		for _, i := range catids {
			cat := GetCat(u.Tostr(i))
			allCats = append(allCats, cat)
		}
	}
	fmt.Printf("GetCategories: %v\n", time.Now().Sub(t0))
	return allCats
}
func GetAllCategories() []Category {
	var allCats []Category
	var catids []int
	err := mc.Get("CategoryList", &catids)
	if err != nil {
		rows, err := stmtGetAllCats.Query()
		if err != nil {
			glog.Errorf("stmtGetAllCats: %s", err)
			return allCats
		}
		for rows.Next() {
			var cat Category
			rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID, &cat.Exclude)
			allCats = append(allCats, cat)
			catids = append(catids, cat.ID)
			mc.Set("Category"+u.Tostr(cat.ID)+"_", cat)
		}
		mc.Set("CategoryList", allCats)
	} else {
		for _, i := range catids {
			cat := GetCat(u.Tostr(i))
			allCats = append(allCats, cat)
		}
	}
	return allCats
}
func CacheAllCats() {
	for _, c := range GetAllCategories() {
		err := mc.Set("Category"+u.Tostr(c.ID)+"_", c)
		if err != nil {
			glog.Errorf("mc.Set(Category %v _: %s", c.ID, err)
		}
		err = mc.Set("Category"+u.Tostr(c.ID)+"_UnreadCount", c.Unread())
		if err != nil {
			glog.Errorf("mc.Set(Category %v _UnreadCount): %s", c.ID, err)
		}
	}
}
