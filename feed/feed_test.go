package feed

import (
	//	"database/sql"
	//	"fmt"
	//	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	//	_ "github.com/go-sql-driver/mysql"
	//	"github.com/golang/glog"
	"fmt"
	"testing"
	//	"github.com/golang/glog"
)

func TestFeed_DecrementUnread(t *testing.T) {
	print("\tFeed.DecrementUnread()\n")
	seed()
	f := ff()
	fu := f.Unread()
	f.DecrementUnread()
	if f.Unread() != fu-1 {
		t.Errorf("f.Unread len %v <=> %v", fu-1, f.Unread())
	}
}
func TestFeed_IncrementUnread(t *testing.T) {
	print("\tFeed.IncrementUnread()\n")
	seed()
	f := ff()
	fu := f.Unread()
	f.IncrementUnread()
	if f.Unread() != fu+1 {
		t.Errorf("f.Unread len %v <=> %v", fu+1, f.Unread())
	}
}
func TestFeed_Unread(t *testing.T) {
	print("\tFeed.Unread()\n")
	seed()
	print("\t\tInitial\n")
	f := ff()
	fu := f.Unread()
	if fu != 3 {
		t.Errorf("f.Unread 3 <=> %v", fu)
	}
	print("\t\tMark Read\n")
	e1 := GetEntry("1", "test")
	e1.MarkRead()
	fr := f.Unread()
	if fr != 2 {
		t.Errorf("f.Unread 2 <=> %v", fr)
	}
	print("\t\tMark Unread\n")
	e1.MarkUnread()
	fun := f.Unread()
	if fun != 3 {
		t.Errorf("f.Unread 3 <=> %v", fun)
	}
}
func TestFeed_UnreadEntries(t *testing.T) {
	print("\tFeed.UnreadEntries()\n")
	seed()
	f := ff()
	print("\t\tInitial\n")
	ul := len(f.UnreadEntries())
	if ul != 3 {
		t.Errorf("f.UnreadEntries len 3 <=> %v", ul)
	}
	print("\t\tMark Read\n")
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	if err != nil {
		t.Errorf("e1.MarkRead(): %s", err)
	}
	ul = len(f.UnreadEntries())
	if ul != 2 {
		t.Errorf("f.UnreadEntries len 2 <=> %v", ul)
	}
	print("\t\tMark Unread\n")
	err = e1.MarkUnread()
	if err != nil {
		t.Errorf("e1.MarkUnRead(): %s", err)
	}
	ul = len(f.UnreadEntries())
	if ul != 3 {
		t.Errorf("f.UnreadEntries len 3 <=> %v", ul)
	}

}
func TestFeed_MarkedEntries(t *testing.T) {
	print("\tFeed.MarkedEntries()\n")
	seed()
	f := ff()

	print("\t\tInitial\n")
	ml := len(f.MarkedEntries())
	if ml != 0 {
		t.Errorf("f.MarkedEntries len 0 <=> %v", ml)
	}

	print("\t\tMarking\n")
	e1 := GetEntry("1", "test")
	err := e1.Mark()
	if err != nil {
		t.Errorf("e1.Mark(): %s", err)
	}
	ml = len(f.MarkedEntries())
	if ml != 1 {
		t.Errorf("f.MarkedEntries len 1 <=> %v", ml)
	}

	print("\t\tUnMarking\n")
	err = e1.UnMark()
	if err != nil {
		t.Errorf("e1.UnMark(): %s", err)
	}
	ml = len(f.MarkedEntries())
	if ml != 0 {
		t.Errorf("f.MarkedEntries len 0 <=> %v", ml)
	}
}
func TestFeed_ReadEntries(t *testing.T) {
	print("\tFeed.ReadEntries()\n")
	seed()
	f := ff()

	print("\t\tInitial\n")
	ml := len(f.ReadEntries())
	if ml != 0 {
		t.Errorf("f.ReadEntries len 0 <=> %v", ml)
	}

	print("\t\tMarking Read\n")
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	if err != nil {
		t.Errorf("e1.MarkRead(): %s", err)
	}
	ml = len(f.ReadEntries())
	if ml != 1 {
		t.Errorf("f.ReadEntries len 1 <=> %v", ml)
	}

	print("\t\tUnMarkingRead\n")
	err = e1.MarkUnread()
	if err != nil {
		t.Errorf("e1.MarkUnread(): %s", err)
	}
	ml = len(f.ReadEntries())
	if ml != 0 {
		t.Errorf("f.ReadEntries len 0 <=> %v", ml)
	}
}
func TestFeed_SearchTitles(t *testing.T) {
	print("\tFeed.SearchTitles\n")
	seed()
	f := ff()

	print("\t\tUnread\n")
	ul := len(f.SearchTitles("asdf", "unread"))
	if ul != 1 {
		t.Errorf("f.SearchTitles(asdf,unread) len 1 <=> %v", ul)
	}

	print("\t\tRead, before marking read\n")
	ul = len(f.SearchTitles("asdf", "read"))
	if ul != 0 {
		t.Errorf("f.SearchTitles(asdf,read) len 0 <=> %v", ul)
	}
	print("\t\tRead, after marking read\n")
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	if err != nil {
		t.Errorf("e1.MarkRead(): %s", err)
	}
	ul = len(f.SearchTitles("asdf", "read"))
	if ul != 1 {
		t.Errorf("f.SearchTitles(asdf,read) len 1 <=> %v", ul)
	}

	print("\t\tMarked, before marking\n")
	ul = len(f.SearchTitles("asdf", "marked"))
	if ul != 0 {
		t.Errorf("f.SearchTitles(asdf,marked) len 0 <=> %v", ul)
	}
	print("\t\tMarked, after marking\n")
	err = e1.Mark()
	if err != nil {
		t.Errorf("e1.Mark(): %s", err)
	}
	ul = len(f.SearchTitles("asdf", "marked"))
	if ul != 1 {
		t.Errorf("f.SearchTitles(asdf,marked) len 1 <=> %v", ul)
	}

	print("\t\tAll\n")
	ul = len(f.SearchTitles("asdf", "all"))
	if ul != 1 {
		t.Errorf("f.SearchTitles(asdf,all) len 1 <=> %v", ul)
	}
}
func TestFeed_Excludes(t *testing.T) {
	print("\tFeed.Excludes()\n")
	seed()
	f := ff()

	print("\t\tBefore adding any\n")
	el := len(f.Excludes())
	if el != 0 {
		t.Errorf("f.Excludes() len 0 <=> %v", el)
	}

	print("\t\tAdding excludes\n")
	f.Exclude = "a,b,c,asdf,"
	f.Save()
	f = ff()
	el = len(f.Excludes())
	if el != 4 {
		t.Errorf("f.Excludes() len 4 <=> %v", el)
	}
}
func TestFeed_ExcludesData(t *testing.T) {
	print("\tFeed.ExcludesData()\n")
	seed()
	f := ff()

	print("\t\tInitial\n")
	el := len(f.ExcludesData())
	if el != 0 {
		t.Errorf("f.ExcludesData() len 0 <=> %v", el)
	}

	print("\t\tAdding excludes for data\n")
	f.ExcludeData = "a,b,c,asdf,"
	f.Save()
	f = ff()
	el = len(f.ExcludesData())
	if el != 4 {
		t.Errorf("f.ExcludesData() len 4 <=> %v", el)
	}
}
func TestFeed_GetEntriesByParam(t *testing.T) {
	print("\tFeed.GetEntriesByParam()\n")
	seed()
	f := ff()
	fl := len(f.GetEntriesByParam("id=1"))
	if fl != 1 {
		t.Errorf("f.getEntriesByParam(id=1) len 1 <=> %v", fl)
	}

	fl = len(f.GetEntriesByParam("1=1"))
	if fl != 3 {
		t.Errorf("f.getEntriesByParam(1=1) len 3 <=> %v", fl)
	}
}
func TestFeed_Save(t *testing.T) {
	print("\tFeed.Save()\n")
	seed()
	f := ff()
	print("\t\tModification\n")
	f.Title = "ahoy"
	err := f.Save()
	if err != nil {
		t.Errorf("f.Save(): %s", err)
	}
	g := ff()
	if g.Title != "ahoy" {
		t.Errorf("f.Save, name ahoy <=> %s", g.Title)
	}
	print("\t\tNew\n")
	ifl := len(GetAllFeeds())
	var f2 Feed
	f2.UserName = "newuser"
	f2.Url = "myurl"
	err = f2.Save()
	if err != nil {
		t.Errorf("f2.Save(): %s", err)
	}
	nfl := len(GetAllFeeds())
	if nfl != ifl+1 {
		t.Errorf("Len(GetAllFeeds) did not increase: %v <=> %v", nfl, ifl)
	}
}
func TestFeed_Class(t *testing.T) {
	print("\tFeed.Class()\n")
	seed()
	f := ff()
	print("\t\tInitial\n")
	if f.Class() != "oddUnread" {
		t.Errorf("f.Class() expected oddUnread <=> %s", f.Class())
	}
	print("\t\tAfter Marking Read\n")
	for _, e := range f.UnreadEntries() {
		err := e.MarkRead()
		if err != nil {
			t.Errorf("e.MarkRead(): %s", err)
		}
	}
	if f.Class() != "odd" {
		t.Errorf("f.Class expected odd <=> %s", f.Class())
	}
}
func TestFeed_Category(t *testing.T) {
	print("\tFeed.Category()\n")
	seed()
	f := ff()
	c := f.Category()
	if c.ID != 1 {
		t.Errorf("f.Category().ID 1 <=> %v", c.ID)
	}
}
func TestFeed_ClearEntries(t *testing.T) {
	print("\tFeed.ClearEntries()\n")
	seed()
	f := ff()
	e := GetEntry("1", "test")
	e.MarkRead()
	popcache(f, t)

	f.ClearEntries()
	var el []Entry
	err := mc.Get("Category"+u.Tostr(f.CategoryID)+"_unreadentries", &el)
	if err.Error() != "memcache: cache miss" {
		t.Errorf(" not zeroed")
	}
	err = mc.Get("Category"+u.Tostr(f.CategoryID)+"_readentries", &el)
	if err.Error() != "memcache: cache miss" {
		t.Errorf(" not zeroed")
	}
	err = mc.Get("Feed"+u.Tostr(f.CategoryID)+"_unreadentries", &el)
	if err.Error() != "memcache: cache miss" {
		t.Errorf(" not zeroed")
	}
	err = mc.Get("Feed"+u.Tostr(f.CategoryID)+"_readentries", &el)
	if err.Error() != "memcache: cache miss" {
		t.Errorf(" not zeroed")
	}
}
func TestFeed_ClearMarked(t *testing.T) {
	print("\tFeed.ClearMarked()\n")
	seed()
	f := ff()
	e := GetEntry("1", "test")
	e.Mark()
	var icul []Entry
	_ = f.MarkedEntries()
	_ = f.Category().MarkedEntries()
	err := mc.Get("Feed"+u.Tostr(f.ID)+"_markedentries", &icul)
	if err != nil {
		t.Errorf("Feed"+u.Tostr(f.ID)+"_markedentries: %s", err)
	}
	err = mc.Get("Category"+u.Tostr(f.CategoryID)+"_markedentries", &icul)
	if err != nil {
		t.Errorf("Category"+u.Tostr(f.CategoryID)+"_markedentries: %s", err)
	}

	f.ClearMarked()
	err = mc.Get("Feed"+u.Tostr(f.ID)+"_markedentries", &icul)
	if err.Error() != "memcache: cache miss" {
		t.Errorf("Feed"+u.Tostr(f.ID)+"_markedentries not zeroed: %s", err)
	}
	err = mc.Get("Category"+u.Tostr(f.CategoryID)+"_markedentries", &icul)
	if err.Error() != "memcache: cache miss" {
		t.Errorf("Category"+u.Tostr(f.ID)+"_markedentries not zeroed: %s", err)
	}
}
func TestFeed_DeleteExcludes(t *testing.T) {
	fmt.Print("\tFeed.DeleteExcludes()\n")
	seed()
	f := ff()
	uel := len(f.UnreadEntries())
	f.Exclude = "asdf"
	f.Save()
	f.DeleteExcludes()
	if len(f.UnreadEntries()) == uel {
		t.Errorf("Length of unreadentries did not shrink")
	}
	seed()
	f.ExcludeData = "asdf"
	e := GetEntry("1", "test")
	e.Content = "my asdf is a first name"
	e.Save("test")
	f.DeleteExcludes()
	if len(f.UnreadEntries()) == uel {
		t.Errorf("Length of unreadentries did not shrink")
	}
}
func TestFeed_AllEntries(t *testing.T) {
	print("\tfeed.AllEntries()\n")
	seed()
	f := ff()
	el := len(f.AllEntries())
	if el != 3 {
		t.Errorf("len(feed.AllEntries()) 3 <=> %v", el)
	}

}
func TestFeed_Delete(t *testing.T) {
	print("\tFeed.Delete()\n")
	f := sf()
	afl := len(GetAllFeeds())
	err := f.Delete()
	if err != nil {
		t.Errorf("f.Delete(): %s", err)
	}
	nfl := len(GetAllFeeds())
	if afl == nfl {
		t.Errorf("Length of GetallFeeds did not change")
	}
}
func TestFeed_MarkEntriesRead(t *testing.T) {
	print("\tFeed.MarkEntriesRead()\n")
	f := sf()
	el := []string{"1", "2"}
	err := f.MarkEntriesRead(el)
	if err != nil {
		t.Errorf("c1.markEntriesRead([1,2]): %s", err)
	}
	lue := len(f.UnreadEntries())
	if lue != 2 {
		t.Errorf("f.UnreadEntries() len 2 <=> %v", lue)
	}
}
func TestFeed_ClearCache(t *testing.T) {
	print("\tFeed.ClearCache\n")
	f := sf()
	popcache(f, t)
	cl := []string{"Feed" + u.Tostr(f.ID) + "_",
		"FeedsWithoutCats" + f.UserName,
		"FeedList",
		"Feed" + u.Tostr(f.ID) + "_UnreadCount",
		"Feed" + u.Tostr(f.ID) + "_readentries",
		"Feed" + u.Tostr(f.ID) + "_unreadentries",
		"Feed" + u.Tostr(f.ID) + "_markedentries",
	}
	f.ClearCache()
	for _, i := range cl {
		err := mc.Delete(i)
		if err.Error() != "memcache: cache miss" {
			t.Errorf("mc.Delete(%s): %s", i, err)
		}
	}

}

func popcache(f Feed, t *testing.T) {
	var icul []Entry
	var icl []Entry
	var iful []Entry
	var ifl []Entry
	c := f.Category()
	_ = c.UnreadEntries()
	_ = c.ReadEntries()
	_ = f.UnreadEntries()
	_ = f.ReadEntries()
	err := mc.Get("Category"+u.Tostr(f.CategoryID)+"_unreadentries", &icul)
	if err != nil {
		t.Errorf("mc.get(Category(f.id)_unreadentries): %s", err)
	}

	err = mc.Get("Category"+u.Tostr(c.ID)+"_readentries", &icl)
	if err != nil {
		t.Errorf("mc.Get(Category(c.id)_readentries: %s", err)
	}

	err = mc.Get("Feed"+u.Tostr(f.ID)+"_unreadentries", &iful)
	if err != nil {
		t.Errorf("mc.get(f.id)_unreadentries): %s", err)
	}

	err = mc.Get("Feed"+u.Tostr(f.ID)+"_readentries", &ifl)
	if err != nil {
		t.Errorf("mc.Get(f.id)_readentries: %s", err)
	}
}
func sf() Feed {
	seed()
	return ff()
}
func ff() Feed {
	f, _ := GetFeed(1)
	return f
}
