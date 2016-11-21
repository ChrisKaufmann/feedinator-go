package feed

import (
	"html/template"
	"testing"
	"github.com/stvp/assert"
)

func ef() Entry {
	seed()
	return ge()
}
func ge() Entry {
	return GetEntry("1", "test")
}

func TestGetEntry(t *testing.T) {
	e := ef()
	assert.Equal(t, "asdf", e.Title)
}
func TestEntry_Save(t *testing.T) {
	e := ef()
	e.Title = "newtitle"
	e.Link = "newlink"
	e.Marked = "1"
	e.FeedID = 2
	e.Content = template.HTML("newcontent")
	e.Unread = false
	err := e.Save("test")
	assert.Nil(t, err)
	ne := ge()
	assert.Equal(t, "newtitle", ne.Title)
	assert.Equal(t, "newlink", ne.Link)
	assert.Equal(t, "1", ne.Marked)
	assert.Equal(t, 2, ne.FeedID)
	assert.False(t, ne.Unread)
	if ne.Content != "newcontent" {
		t.Errorf("Content != newcontent: %s", ne.Content)
	}
}
func TestEntry_ViewMode(t *testing.T) {
	e := ef()
	assert.Equal(t,"default", e.ViewMode(), "initial entry.ViewMode")
	f, err := GetFeed(1)
	assert.Nil(t,err,"GetFeed(1)")
	f.ViewMode = "link"
	err = f.Save()
	assert.Nil(t, err)
	assert.Equal(t, "link", e.ViewMode())
}
func TestEntry_AutoscrollPX(t *testing.T) {
	e := ef()
	assert.Equal(t,0,e.AutoscrollPX())
	f, err := GetFeed(1)
	assert.Nil(t, err)
	f.AutoscrollPX = 1234
	err = f.Save()
	assert.Nil(t, err)
	assert.Equal(t,1234,e.AutoscrollPX())
}
func TestEntry_Feed(t *testing.T) {
	e := ef()
	assert.Equal(t,1,e.Feed().ID)
}
func TestEntry_Mark(t *testing.T) {
	e := ef()
	err := e.Mark()
	assert.Nil(t,err)
	e = ge()
	assert.Equal(t,"1",e.Marked)
}
func TestEntry_MarkRead(t *testing.T) {
	e := ef()
	err := e.MarkRead()
	assert.Nil(t,err)
	assert.False(t, ge().Unread)
}
func TestEntry_MarkUnread(t *testing.T) {
	e := ef()
	err := e.MarkRead()
	assert.Nil(t, err)
	e = ge()
	assert.False(t, e.Unread)
	err = e.MarkUnread()
	assert.Nil(t, err)
	e = ge()
	assert.True(t, e.Unread)
}
func TestEntry_Normalize(t *testing.T) {
	e := ef()
	e = e.Normalize()
	assert.Equal(t, "unset", e.MarkSet)
	assert.Equal(t, "unread", e.ReadUnread)
}
func TestEntry_ToggleMark(t *testing.T) {
	e := ef()
	nm, err := e.ToggleMark()
	assert.Nil(t, err)
	assert.Equal(t, "set", nm)
	e = ge()
	assert.Equal(t, "1", e.Marked)
	assert.Nil(t,err)
	nm, err = e.ToggleMark()
	assert.Equal(t, "unset", nm)
	assert.Nil(t,err)
	e = ge()
	assert.Equal(t, "0", e.Marked)
}
func TestEntry_UnMark(t *testing.T) {
	e := ef()
	err := e.Mark()
	assert.Nil(t,err)
	e = ge()
	assert.Equal(t, "1", e.Marked)
	err = e.UnMark()
	assert.Nil(t,err)
	e = ge()
	assert.Equal(t, "0", e.Marked)
}
func TestGetEntriesCount(t *testing.T) {
	_ = ef()
	ec, err := GetEntriesCount()
	assert.Nil(t,err)
	assert.Equal(t, "5", ec)
}
func TestAllMarkedEntries(t *testing.T) {
	e := ef()
	assert.Equal(t,0, len(AllMarkedEntries("test")))
	err := e.Mark()
	assert.Nil(t,err)
	assert.Equal(t, 1, len(AllMarkedEntries("test")))
}

/*func TestEntry_ProxyLink(t *testing.T) {
	e := ef()
	e.Link = "http://127.0.0.1:9001/"
	c, err := e.ProxyLink()
	if err != nil {
		t.Errorf("e.Proxylink: %s", err)
	}
	fmt.Printf("content: %s", c)
	os.Exit(1)
}
*/
