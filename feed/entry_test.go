package feed

import (
	"testing"
	"html/template"
	"fmt"
)
func ef() ( Entry) {
	seed()
	return ge()
}
func ge()  Entry {
	return GetEntry("1","test")
}

func TestEntry_GetEntry(t *testing.T) {
	fmt.Print("\tGetEntry()\n")
	e := ef()
	if e.Title != "asdf"{
		t.Errorf("e.Title: asdf <=> %s", e.Title)
	}
}
func TestEntry_Save(t *testing.T) {
	fmt.Print("\tEntry.Save()\n")
	e := ef()
	e.Title="newtitle"
	e.Link="newlink"
	e.Marked="1"
	e.FeedID=2
	e.Content=template.HTML("newcontent")
	e.Unread=false
	err := e.Save("test")
	if err != nil {
		t.Errorf("Error saving: %s",err)
	}
	ne := ge()
	if ne.Title != "newtitle" {t.Errorf("title != newtitle: %s",ne.Title)}
	if ne.Link != "newlink"	{t.Errorf("link != newlink: %s", ne.Link)}
	if ne.Marked != "1" {t.Errorf("marked != 1: %s", ne.Marked)}
	if ne.FeedID != 2 {t.Errorf("FeedID != 2: %v", ne.FeedID)}
	if ne.Content != "newcontent" {t.Errorf("Content != newcontent: %s", ne.Content)}
	if ne.Unread != false {t.Errorf("Unread != false: %t",ne.Unread)}
}
func TestEntry_ViewMode(t *testing.T) {
	print("\tEntry.ViewMode()\n")
	e := ef()
	print("\t\tInitial\n")
	if e.ViewMode() != "" {
		t.Errorf("e.Viewmode: oddUnread<=>%s",e.ViewMode())
	}
	print("\t\tUpdated\n")
	f,err := GetFeed(1)
	if err != nil {
		t.Errorf("GetFeed(1): %s",err)
	}
	f.ViewMode = "link"
	err = f.Save()
	if err != nil {
		t.Errorf("f.Save(): %s", err)
	}
	if e.ViewMode() != "link" {
		t.Errorf("e.ViewMode: link<=>%s",e.ViewMode())
	}
}
func TestEntry_AutoscrollPX(t *testing.T) {
	print("\tEntry.AutoscrollPX()\n")
	e:=ef()
	print("\t\tInitial\n")
	if e.AutoscrollPX() != 0 {
		t.Errorf("e.AutoscrollPX: 0<=>%v",e.AutoscrollPX())
	}
	print("\t\tUpdated\n")
	f,err := GetFeed(1)
	if err != nil {
		t.Errorf("GetFeed(1): %s",err)
	}
	f.AutoscrollPX = 1234
	err = f.Save()
	if err != nil {
		t.Errorf("f.Save(): %s", err)
	}
	if e.AutoscrollPX() != 1234 {
		t.Errorf("e.AutoscrollPX(): 1234 <=> %v",e.AutoscrollPX())
	}
}
func TestEntry_Feed(t *testing.T) {
	print("\tEntry.Feed()\n")
	e:=ef()
	if e.Feed().ID != 1 {
		t.Errorf("e.Feed().ID 1 <=> %v",e.Feed().ID)
	}
}
func TestEntry_Mark(t *testing.T) {
	print("\tEntry.Mark()\n")
	e:=ef()
	err := e.Mark()
	if err != nil {
		t.Errorf("e.Mark(): %s", err)
	}
	e = ge()
	if e.Marked != "1" {
		t.Errorf("e.Marked: 1 <=> %s", e.Marked)
	}

}
func TestEntry_MarkRead(t *testing.T) {
	print("\tEntry.MarkRead()\n")
	e:=ef()
	err := e.MarkRead()
	if err != nil {
		t.Errorf("e.MarkRead(): %s", err)
	}
	e=ge()
	if e.Unread != false {
		t.Errorf("e.Unread: false <=> %t",e.Unread)
	}
}
func TestEntry_MarkUnread(t *testing.T) {
	print("\tEntry.MarkUnread()\n")
	e := ef()
	print("\t\tMarking Read first\n")
	err := e.MarkRead()
	if err != nil {
		t.Errorf("e.MarkREad(): %s", err)
	}
	e=ge()
	if e.Unread != false {
		t.Errorf("e.Unread: false <=> %t", e.Unread)
	}
	print("\t\tMarking Unread\n")
	err = e.MarkUnread()
	if err != nil {
		t.Errorf("e.MarkUnread(): %s", err)
	}
	e = ge()
	if e.Unread != true {
		t.Errorf("e.Unread: true <=> %t", e.Unread)
	}

}
func TestEntry_Normalize(t *testing.T) {
	print("\tEntry.Normalize()\n")
	e:=ef()
	e=e.Normalize()
	if e.MarkSet != "unset" {
		t.Errorf("e.MarkSet unset <=> %s",e.MarkSet)
	}
	if e.ReadUnread != "unread" {
		t.Errorf("e.ReadUnread unread <=> %s", e.ReadUnread)
	}
}
func TestEntry_ToggleMark(t *testing.T) {
	print("\tEntry.ToggleMark\n")
	e:=ef()
	print("\t\tToggling to marked")
	nm, err := e.ToggleMark()
	if nm != "set" {
		t.Errorf("ToggleMark returned: set <=> %s", nm)
	}
	e = ge()
	if e.Marked != "1" {
		t.Errorf("e.Marked: 1 <=> %s", e.Marked)
	}
	if err != nil {
		t.Errorf("e.ToggleMark(): %s", err)
	}
	print("\t\tToggling to Unmarked")
	nm, err = e.ToggleMark()
	if nm != "unset" {
		t.Errorf("Togglemark returned: unset <=> %s", nm)
	}
	if err != nil {
		t.Errorf("e.ToggleMark(): %s", err)
	}
	e = ge()
	if e.Marked != "0" {
		t.Errorf("e.Marked: 0 <=> %s", e.Marked)
	}
}
func TestEntry_UnMark(t *testing.T) {
	print("\tEntry.UnMark()\n")
	e := ef()
	print("\t\tInitial\n")
	err := e.Mark()
	if err != nil {
		t.Errorf("e.Mark(): %s", err)
	}
	e=ge()
	if e.Marked != "1" {
		t.Errorf("e.Marked: 1 <=> %s", e.Marked)
	}
	err = e.UnMark()
	if err != nil {
		t.Errorf("e.UnMark(): %s", err)
	}
	e = ge()
	if e.Marked != "0" {
		t.Errorf("e.Marked: 0 <=> %s", e.Marked)
	}
}
func TestGetEntriesCount(t *testing.T) {
	print("\tEntry: GetEntriesCount()\n")
	_ =ef()
	ec,err := GetEntriesCount()
	if err != nil {
		t.Errorf("GetEntriesCount(): %s", err)
	}
	if ec != "5" {
		t.Errorf("GetEntriesCount: count: 5 <=> %s", ec)
	}
}