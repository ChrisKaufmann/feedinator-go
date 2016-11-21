package feed

import (
	//	"database/sql"
	//	"fmt"
	//	"github.com/ChrisKaufmann/easymemcache"
	u "github.com/ChrisKaufmann/goutils"
	//	_ "github.com/go-sql-driver/mysql"
	//	"github.com/golang/glog"
	//"fmt"
	"testing"
	//	"github.com/golang/glog"
	"github.com/stvp/assert"
)

func TestFeed_DecrementUnread(t *testing.T) {
	seed()
	f := ff()
	fu := f.Unread()
	f.DecrementUnread()
	assert.Equal(t, fu-1, f.Unread())
}
func TestFeed_IncrementUnread(t *testing.T) {
	seed()
	f := ff()
	fu := f.Unread()
	f.IncrementUnread()
	assert.Equal(t, fu+1, f.Unread())
}
func TestFeed_Unread(t *testing.T) {
	seed()
	f := ff()
	assert.Equal(t,3,f.Unread())
	e1 := GetEntry("1", "test")
	e1.MarkRead()
	assert.Equal(t,2,f.Unread())
	e1.MarkUnread()
	assert.Equal(t,3,f.Unread())
}
func TestFeed_UnreadEntries(t *testing.T) {
	seed()
	f := ff()
	assert.Equal(t,3,len(f.UnreadEntries()))
	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	assert.Nil(t,err,"e1.MarkRead()")
	assert.Equal(t,2,len(f.UnreadEntries()))
	err = e1.MarkUnread()
	assert.Nil(t, err, "e1.MarkUnread()")
	assert.Equal(t,3,len(f.UnreadEntries()))

}
func TestFeed_MarkedEntries(t *testing.T) {
	seed()
	f := ff()
	assert.Equal(t,0,len(f.MarkedEntries()))
	e1 := GetEntry("1", "test")
	err := e1.Mark()
	assert.Nil(t,err,"e1.Mark()")
	assert.Equal(t,1,len(f.MarkedEntries()))
	err = e1.UnMark()
	assert.Nil(t, err, "e1.UnMark()")
	assert.Equal(t,0,len(f.MarkedEntries()))
}
func TestFeed_ReadEntries(t *testing.T) {
	seed()
	f := ff()
	assert.Equal(t,0,len(f.ReadEntries()))

	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	assert.Nil(t,err,"e1.MarkRead()")
	assert.Equal(t,1,len(f.ReadEntries()))

	err = e1.MarkUnread()
	assert.Nil(t,err,"e1.MarkUnread()")
	assert.Equal(t,0,len(f.ReadEntries()))
}
func TestFeed_SearchTitles(t *testing.T) {
	seed()
	f := ff()

	assert.Equal(t,1,len(f.SearchTitles("asdf", "unread")))
	assert.Equal(t,0,len(f.SearchTitles("asdf", "read")))

	e1 := GetEntry("1", "test")
	err := e1.MarkRead()
	assert.Nil(t,err,"e1.MarkRead()")
	assert.Equal(t,1,len(f.SearchTitles("asdf", "read")))

	assert.Equal(t,0,len(f.SearchTitles("asdf", "marked")))
	err = e1.Mark()
	assert.Nil(t,err,"e1.Mark()")
	assert.Equal(t,1,len(f.SearchTitles("asdf", "marked")))

	assert.Equal(t,1,len(f.SearchTitles("asdf", "all")))
}
func TestFeed_Excludes(t *testing.T) {
	seed()
	f := ff()

	assert.Equal(t,0,len(f.Excludes()))
	f.Exclude = "a,b,c,asdf,"
	f.Save()
	f = ff()
	assert.Equal(t,4,len(f.Excludes()))
}
func TestFeed_ExcludesData(t *testing.T) {
	seed()
	f := ff()

	assert.Equal(t,0,len(f.ExcludesData()))

	f.ExcludeData = "a,b,c,asdf,"
	f.Save()
	f = ff()
	assert.Equal(t,4,len(f.ExcludesData()))
}
func TestFeed_GetEntriesByParam(t *testing.T) {
	seed()
	f := ff()
	assert.Equal(t, 1,len(f.GetEntriesByParam("id=1")))

	assert.Equal(t,3, len(f.GetEntriesByParam("1=1")))
}
func TestFeed_Save(t *testing.T) {
	seed()
	f := ff()
	f.Title = "ahoy"
	err := f.Save()
	assert.Nil(t,err)
	g := ff()
	assert.Equal(t, "ahoy", g.Title)

	ifl := len(GetAllFeeds())
	var f2 Feed
	f2.UserName = "newuser"
	f2.Url = "myurl"
	err = f2.Save()
	assert.Nil(t,err)
	assert.Equal(t,len(GetAllFeeds()), ifl+1)
	nfl := len(GetAllFeeds())
	assert.Equal(t, nfl, ifl+1)
}
func TestFeed_Class(t *testing.T) {
	seed()
	f := ff()
	assert.Equal(t, "oddUnread", f.Class())
	for _, e := range f.UnreadEntries() {
		err := e.MarkRead()
		assert.Nil(t, err)
	}
	assert.Equal(t, "odd", f.Class())
}
func TestFeed_Category(t *testing.T) {
	seed()
	f := ff()
	c := f.Category()
	assert.Equal(t, 1, c.ID)
}
func TestFeed_ClearEntries(t *testing.T) {
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
	seed()
	assert.Equal(t, 3, len(ff().AllEntries()))
}
func TestFeed_Delete(t *testing.T) {
	f := sf()
	afl := len(GetAllFeeds())
	err := f.Delete()
	assert.Nil(t, err)
	assert.NotEqual(t, afl, len(GetAllFeeds()))
}
func TestFeed_MarkEntriesRead(t *testing.T) {
	f := sf()
	el := []string{"1", "2"}
	err := f.MarkEntriesRead(el)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(f.UnreadEntries()))
}
func TestFeed_ClearCache(t *testing.T) {
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
func TestFeed_HasError(t *testing.T) {
	f := sf()
	assert.False(t, f.HasError())
	f.ErrorString = "new error"
	f.Save()
	g := ff()
	assert.True(t, g.HasError())
}

func TestGetFeedsWithoutCats(t *testing.T) {
	seed()
	assert.Equal(t, 2, len(GetFeedsWithoutCats("test")))
	f, err := GetFeed(5)
	assert.Nil(t, err)
	f.CategoryID = 1
	err = f.Save()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(GetFeedsWithoutCats("test")))
}
func TestGetFeeds(t *testing.T) {
	seed()
	assert.Equal(t, 4, len(GetFeeds("test")))
}
func TestCacheAllFeeds(t *testing.T) {
	seed()
	var f Feed
	err := mc.Get("Feed1_", &f)
	assert.NotNil(t,err)
	CacheAllFeeds()
	err = mc.Get("Feed1_", &f)
	assert.Nil(t, err)
	var uc int
	err = mc.Get("Feed1_UnreadCount", &uc)
	assert.Nil(t, err)
}
func TestGetAllFeeds(t *testing.T) {
	seed()
	assert.Equal(t, 8, len(GetAllFeeds()))
}
func TestGetCategoryFeeds(t *testing.T) {
	seed()
	assert.Equal(t,2,len(GetCategoryFeeds(1)))
}
func TestGetFeed(t *testing.T) {
	seed()
	f, err := GetFeed(1)
	assert.Nil(t,err)
	assert.Equal(t, "test1", f.Title)
}
func TestFeed_SkippableEntry(t *testing.T) {
	var e Entry
	e.Title = "aaAAaaAA"
	var f Feed
	f.Exclude = "abababab"
	tf, err := f.SkippableEntry(e)
	assert.Nil(t, err, "SkippableEntry")
	assert.False(t, tf, "f.SkippableEntry()")
	f.Exclude = "abababab,aaa"
	tf, err = f.SkippableEntry(e)
	assert.Nil(t, err, "SkippableEntry")
	assert.True(t, tf, "f.SkippableEntry()")

}
func TestFeed_Update(t *testing.T) {
/*
	rss_string := `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" media="screen" href="/~d/styles/rss2full.xsl"?><?xml-stylesheet type="text/css" media="screen" href="http://feeds.feedburner.com/~d/styles/itemcontent.css"?><rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:feedburner="http://rssnamespace.org/feedburner/ext/1.0" version="2.0">
  <channel>
    <title>Slickdeals Frontpage RSS Feed</title>
    <link>https://slickdeals.net/</link>
    <description>
              Save  money with Slickdeals: find the lowest and cheapest price, best deals  and bargains, and hot coupons. Community driven bargain hunting with thousands of free discounts, promo codes, freebies and price  comparisons.
          </description>
    <language>en</language>
    <lastBuildDate>Fri, 18 Nov 2016 21:27:23 GMT</lastBuildDate>
    <ttl>5</ttl>

    <image>
      <url>https://static.slickdealscdn.com/images/misc/rss.jpg</url>
      <title>Slickdeals Frontpage RSS Feed</title>
      <link>https://slickdeals.net/</link>
    </image>
    <copyright>Copyright 1999 - 2016</copyright>
          <atom10:link xmlns:atom10="http://www.w3.org/2005/Atom" rel="self" type="application/rss+xml" href="http://feeds.feedburner.com/SlickdealsnetFP" /><feedburner:info uri="slickdealsnetfp" /><atom10:link xmlns:atom10="http://www.w3.org/2005/Atom" rel="hub" href="http://pubsubhubbub.appspot.com/" /><item>
                                            <title>Laika 3D Blu-ray Boxset: Paranorman + Coraline + Boxtrolls (Region Free) $14.50 Shipped</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/eIdmm5xkxSk/9365459-laika-3d-blu-ray-boxset-paranorman-coraline-boxtrolls-region-free-14-50-shipped</link>
                                                <description>Amazon.co.uk has *Laika Boxset: Paranorman + Coraline + Boxtrolls *(Region Free 3D Blu-ray) on sale for £8.16 (price drops in checkout) + £3.08 shipping = £11.24 or *$14.47 USD*. Thanks persian_mafia...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/8/2/3/7/7/200x200/4832391.thumb" alt="Laika 3D Blu-ray Boxset: Paranorman + Coraline + Boxtrolls (Region Free) $14.50 Shipped"></div><br /><div>Thumb Score: +21 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=54b7dc96add111e68ad8fead90eb2890&amp;au=5b37561e763a11e2b02f02e30470827f&amp;pno=236859&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Amazon+co+uk&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9365459" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Amazon.co.uk</a> has <b>Laika Boxset: Paranorman + Coraline + Boxtrolls </b>(Region Free 3D Blu-ray) on sale for £8.16 (price drops in checkout) + £3.08 shipping = £11.24 or <b>$14.47 USD</b>. Thanks persian_mafia<br />
<br />
Includes:<br />
<ul><li>Paranorman (3D Blu-ray)</li>
<li>Coraline  (3D Blu-ray)</li>
<li>Boxtrolls  (3D Blu-ray)</li>
</ul></div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/eIdmm5xkxSk" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 20:55:22 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>persian_mafia</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9365459-laika-3d-blu-ray-boxset-paranorman-coraline-boxtrolls-region-free-14-50-shipped?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9365459-laika-3d-blu-ray-boxset-paranorman-coraline-boxtrolls-region-free-14-50-shipped?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Graco Aire3 Click Connect 3-Wheel Stroller Travel System (Gotham)  $193 + Free Shipping</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/Ona6jrLsgLE/9368299-graco-aire3-click-connect-3-wheel-stroller-travel-system-gotham-193-free-shipping</link>
                                                <description>Amazon.com has *Graco Aire3 Click Connect 3-Wheel Stroller Travel System* (Gotham) on sale for *$192.74*. *Shipping is free*. Thanks klkris

Note, includes Graco Aire3 Click Connect 3-Wheel...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//1/2/2/6/9/8/5/200x200/4832435.thumb" alt="Graco Aire3 Click Connect 3-Wheel Stroller Travel System (Gotham)  $193 + Free Shipping"></div><br /><div>Thumb Score: +19 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=617060e2add211e6a55efead90eb2890&amp;au=bc8574d68c2f11e29f3b6ee3110916cf&amp;pno=236851&amp;lno=1&amp;sdfib=1&amp;afsrc=1&amp;mon=1&amp;trd=Amazon+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9368299" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Amazon.com</a> has <b>Graco Aire3 Click Connect 3-Wheel Stroller Travel System</b> (Gotham) on sale for <b>$192.74</b>. <b><font color="#006400">Shipping is free</font></b>. Thanks klkris <br />
<br />
Note, includes Graco Aire3 Click Connect 3-Wheel Stroller &amp; Graco SnugRide Click Connect 35 Infant Car Seat<br />
<br />
<b>Deal Editor's Notes &amp; Price Research:</b> To learn more about the features of the Aire3 Click Connect 3-Wheel Stroller Travel System, check out this <a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=617060e2add211e6a55efead90eb2890&amp;au=bc8574d68c2f11e29f3b6ee3110916cf&amp;pno=236851&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=YouTube+Video&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9368299" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">YouTube Video</a></div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/Ona6jrLsgLE" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 20:54:43 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>klkris</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9368299-graco-aire3-click-connect-3-wheel-stroller-travel-system-gotham-193-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9368299-graco-aire3-click-connect-3-wheel-stroller-travel-system-gotham-193-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Amazon has Aqueon Betta Pellets Betta Food for $1.18 (Or $0.95 W/ 5+ Subscriptions) And MANY More Options Available For Cheap</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/_5VpUtAKjJo/9338071-amazon-has-aqueon-betta-pellets-betta-food-for-1-18-or-0-95-w-5-subscriptions-and-many-more-options-available-for-cheap</link>
                                                <description><![CDATA[Amazon.com has *Great Prices on Aqueon Fish Supplies* after 'clipped' $1  off coupon (clipped automatically through link    when logged in to Amazon) and checking out via Subscribe & Save. *Shipping...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//6/3/5/4/7/1/200x200/4832371.thumb" alt="Amazon has Aqueon Betta Pellets Betta Food for $1.18 (Or $0.95 W/ 5+ Subscriptions) And MANY More Options Available For Cheap"></div><br /><div>Thumb Score: +32 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=2311931cadd111e6bfc71639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236843&amp;lno=1&amp;sdfib=1&amp;afsrc=1&amp;mon=1&amp;trd=Amazon+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9338071" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Amazon.com</a> has <b>Great Prices on Aqueon Fish Supplies</b> after 'clipped' $1  off coupon (clipped automatically through link    when logged in to Amazon) and checking out via Subscribe &amp; Save. <b><font color="#006400">Shipping is free</font></b>. Thanks jeeves10<br />
<br />
Note, must be logged into your account. Coupons are typically one use  per account.  You may cancel your Subscribe &amp; Save subscription any  time after  your order ships.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/_5VpUtAKjJo" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 20:53:25 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>jeeves10</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9338071-amazon-has-aqueon-betta-pellets-betta-food-for-1-18-or-0-95-w-5-subscriptions-and-many-more-options-available-for-cheap?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9338071-amazon-has-aqueon-betta-pellets-betta-food-for-1-18-or-0-95-w-5-subscriptions-and-many-more-options-available-for-cheap?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Oneida Extra 20% Off Sale: 5-Pc Baking Pan Set $17, 4-Pc SS Canister Set  $14.40 & More + Free S&H]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/V5IPKD7cPdw/9368051-oneida-extra-20-off-sale-5-pc-baking-pan-set-17-4-pc-ss-canister-set-14-40-more-free-s-h</link>
                                                <description>Oneida is having their *Black Friday Preview Sale* and offers an *Additional 20% Off* already reduced prices.  *Shipping is free *with promo code SLICKFREE.  Thanks brisar

Note, prices below after...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//3/8/0/3/8/5/8/200x200/4832347.thumb" alt="Oneida Extra 20% Off Sale: 5-Pc Baking Pan Set $17, 4-Pc SS Canister Set  $14.40 &amp; More + Free S&amp;H"></div><br /><div>Thumb Score: +24 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=f1b6c026add011e6bd891639df3e8759&amp;au=d643819c19b911e499b026acd78395e9&amp;pno=236847&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Oneida&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9368051" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Oneida</a> is having their <b>Black Friday Preview Sale</b> and offers an <b>Additional 20% Off</b> already reduced prices.  <font color="#006400"><b>Shipping is free </b></font>with promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='848823'><strong>SLICKFREE</strong></span></span>.  Thanks brisar<br />
<br />
Note, prices below after promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='848823'><strong>SLICKFREE</strong></span></span> applied in-cart.[LIST][*]<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=f1b6c026add011e6bd891639df3e8759&amp;au=d643819c19b911e499b026acd78395e9&amp;pno=236847&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=4+Piece+Stainless+Steel+Palladia&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9368051" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">4-Piece Stainless Steel Palladian Canister Set w/ Window by Anchor Hocking</a> (pictured) <b>$14.39</b></div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/V5IPKD7cPdw" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 20:50:12 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>brisar</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9368051-oneida-extra-20-off-sale-5-pc-baking-pan-set-17-4-pc-ss-canister-set-14-40-more-free-s-h?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9368051-oneida-extra-20-off-sale-5-pc-baking-pan-set-17-4-pc-ss-canister-set-14-40-more-free-s-h?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Earth's Best Organic Stage 2, Sweet Potato & Beets, 3.5 Ounce Pouch (Pack of 12)  $7.49 or less + free ship]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/U4GcpkjxH0o/9365867-earth-s-best-organic-stage-2-sweet-potato-beets-3-5-ounce-pouch-pack-of-12-7-49-or-less-free-ship</link>
                                                <description><![CDATA[Amazon.com has *12-Count 3.5oz Earth's Best Organic Stage 2 *(Sweet Potato & Beets) on sale for *$7.49* after clipping 20% off coupon and checking out  via Subscribe & Save. *Shipping is free*....]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/8/2/3/7/7/200x200/4832207.thumb" alt="Earth's Best Organic Stage 2, Sweet Potato &amp; Beets, 3.5 Ounce Pouch (Pack of 12)  $7.49 or less + free ship"></div><br /><div>Thumb Score: +22 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=c279198cadce11e69379fead90eb2890&amp;au=5b37561e763a11e2b02f02e30470827f&amp;pno=236839&amp;lno=1&amp;sdfib=1&amp;afsrc=1&amp;mon=1&amp;trd=Amazon+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9365867" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Amazon.com</a> has <b>12-Count 3.5oz Earth's Best Organic Stage 2 </b>(Sweet Potato &amp; Beets) on sale for <b>$7.49</b> after clipping 20% off coupon and checking out  via Subscribe &amp; Save. <b><font color="#006400">Shipping is free</font></b>. Thanks rose2012 <br />
<br />
Note, Prime Members save 15%. You may cancel your Subscribe &amp; Save subscription any time after your order ships.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/U4GcpkjxH0o" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 20:36:52 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>rose2012</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9365867-earth-s-best-organic-stage-2-sweet-potato-beets-3-5-ounce-pouch-pack-of-12-7-49-or-less-free-ship?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9365867-earth-s-best-organic-stage-2-sweet-potato-beets-3-5-ounce-pouch-pack-of-12-7-49-or-less-free-ship?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Asus Q304 2-in-1 13.3" Touchscreen Laptop: Intel Core i5-7200U, 6GB DDR4, 1TB HDD, Win 10 $479.99 + Free Shipping]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/PGizChbRvcw/9366035-asus-q304-2-in-1-13-3-touchscreen-laptop-intel-core-i5-7200u-6gb-ddr4-1tb-hdd-win-10-479-99-free-shipping</link>
                                                <description><![CDATA[BestBuy.com has *ASUS Q304 2-in-1 13.3" Touchscreen Laptop* (Q304UA-BHI5T11)on sale for* $479.99*. *Shipping is free*. Thanks longmanj9

Specs:[LIST][*]Intel Core i5-7200U CPU (7th-Gen Kaby...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4832027.thumb" alt="Asus Q304 2-in-1 13.3&quot; Touchscreen Laptop: Intel Core i5-7200U, 6GB DDR4, 1TB HDD, Win 10 $479.99 + Free Shipping"></div><br /><div>Thumb Score: +29 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=54b2ce0eadcc11e6bfc71639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236835&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=BestBuy+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366035" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">BestBuy.com</a> has <b>ASUS Q304 2-in-1 13.3&quot; Touchscreen Laptop</b> (Q304UA-BHI5T11)on sale for<b> $479.99</b>. <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks longmanj9<br />
<br />
Specs:[LIST][*]Intel Core i5-7200U CPU (7th-Gen Kaby Lake)[*]13.3&quot; 1920x1080 LED Touchscreen Display[*]6GB 2133MHz DDR4 SDRAM[*]1TB 5400RPM Hard Drive[*]Intel HD Graphics 620</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/PGizChbRvcw" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 20:19:28 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>longmanj9</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366035-asus-q304-2-in-1-13-3-touchscreen-laptop-intel-core-i5-7200u-6gb-ddr4-1tb-hdd-win-10-479-99-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366035-asus-q304-2-in-1-13-3-touchscreen-laptop-intel-core-i5-7200u-6gb-ddr4-1tb-hdd-win-10-479-99-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Fandango Buy 2+ Tickets, Get One Free with Visa Checkout for 11/18-11/20/2016</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/6q0z-yGA2RY/9366551-fandango-buy-2-tickets-get-one-free-with-visa-checkout-for-11-18-11-20-2016</link>
                                                <description>Fandango.com has *Tickets: Buy One Get O**ne Free* (limit one ticket free) when you enter promo code DEALSTHATCLICK2 and checkout using Visa Checkout.  Thanks vmanne

Note, you do not need a Visa...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//6/3/5/4/7/1/200x200/4831827.thumb" alt="Fandango Buy 2+ Tickets, Get One Free with Visa Checkout for 11/18-11/20/2016"></div><br /><div>Thumb Score: +110 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=a08c8430adc911e6a1aa1639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236831&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Fandango+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366551" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Fandango.com</a> has <b>Tickets: Buy One Get O</b><b>ne Free</b> (limit one ticket free) when you enter promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='997551'><strong>DEALSTHATCLICK2</strong></span></span> and checkout using Visa Checkout.  Thanks vmanne<br />
<br />
Note, you do not need a Visa card to use Visa Checkout. Convenience fee may apply on the ticket purchased. Free ticket amount will vary, and some exclusions may apply.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/6q0z-yGA2RY" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 19:59:50 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>vmanne</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366551-fandango-buy-2-tickets-get-one-free-with-visa-checkout-for-11-18-11-20-2016?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366551-fandango-buy-2-tickets-get-one-free-with-visa-checkout-for-11-18-11-20-2016?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Walmart: Better Life Technology Garage Floor Cover For $129.98 w/ fs</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/nryb6twfYKo/9291619-walmart-better-life-technology-garage-floor-cover-for-129-98-w-fs</link>
                                                <description><![CDATA[Walmart.com has* 17' x 7.5' Better Life Technology Garage Floor Cover*  on sale for *$129.98*. *Shipping is free*. Thanks ---PerforatedLine---

*Deal Editor's Notes & Price Research:* Provides a...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//6/3/5/4/7/1/200x200/4831231.thumb" alt="Walmart: Better Life Technology Garage Floor Cover For $129.98 w/ fs"></div><br /><div>Thumb Score: +31 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=d9a70968adc211e6afac1639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236823&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Walmart+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9291619" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Walmart.com</a> has<b> 17' x 7.5' Better Life Technology Garage Floor Cover</b>  on sale for <b>$129.98</b>. <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks ---PerforatedLine---<br />
<br />
<b>Deal Editor's Notes &amp; Price Research:</b> Provides a moisture barrier and is standard grade thickness (approximately .055&quot;) - daisybeetle</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/nryb6twfYKo" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 19:11:41 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>---PerforatedLine---</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9291619-walmart-better-life-technology-garage-floor-cover-for-129-98-w-fs?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9291619-walmart-better-life-technology-garage-floor-cover-for-129-98-w-fs?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>GreenWorks Pro GLM801600 80V 21-Inch Cordless Lawn Mower (without battery or charger) $210.83 + FS with prime @ Amazon</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/oseLjHGG_-c/9365511-greenworks-pro-glm801600-80v-21-inch-cordless-lawn-mower-without-battery-or-charger-210-83-fs-with-prime-amazon</link>
                                                <description><![CDATA[Amazon.com has *GreenWorks Pro GLM801600 80V 21" Cordless Lawn Mower* (Battery & Charger Not Included) for *$210.83*.  *Shipping is free*. Thanks kmand20

*Deal Editor's Notes & Price Research:*...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4830807.thumb" alt="GreenWorks Pro GLM801600 80V 21-Inch Cordless Lawn Mower (without battery or charger) $210.83 + FS with prime @ Amazon"></div><br /><div>Thumb Score: +28 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=5b70a1c8adbb11e6bfc71639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236819&amp;lno=1&amp;sdfib=1&amp;afsrc=1&amp;mon=1&amp;trd=Amazon+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9365511" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Amazon.com</a> has <b>GreenWorks Pro GLM801600 80V 21&quot; Cordless Lawn Mower</b> (Battery &amp; Charger Not Included) for <b>$210.83</b>.  <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks kmand20<br />
<br />
<b>Deal Editor's Notes &amp; Price Research:</b> This was part of a Front Page <a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=5b70a1c8adbb11e6bfc71639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236819&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=Deal&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9365511" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Deal</a> a few months ago when it was $233 ($22 more than today's price). - brisar</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/oseLjHGG_-c" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 18:18:02 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>kmand20</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9365511-greenworks-pro-glm801600-80v-21-inch-cordless-lawn-mower-without-battery-or-charger-210-83-fs-with-prime-amazon?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9365511-greenworks-pro-glm801600-80v-21-inch-cordless-lawn-mower-without-battery-or-charger-210-83-fs-with-prime-amazon?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[2x 24-Cans 3 oz Purina Fancy Feast Variety Pack + $5 Target GC for $22.78 w/ 5% S&S or B&M: 30% off w/ cartwheel + $5 Target GC for $16.78]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/G5o2g_gJr1E/9328999-2x-24-cans-3-oz-purina-fancy-feast-variety-pack-5-target-gc-for-22-78-w-5-s-s-or-b-m-30-off-w-cartwheel-5-target-gc-for-16-78</link>
                                                <description><![CDATA[Target has *Great Deals* on *2x 24-Ct 3 oz Purina Fancy Feast Cat Food Variety Packs *when you follow the instructions below. *Shipping is free*. Thanks hjc985, Silvery79 & Waynez...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//6/3/5/4/7/1/200x200/4830107.thumb" alt="2x 24-Cans 3 oz Purina Fancy Feast Variety Pack + $5 Target GC for $22.78 w/ 5% S&amp;S or B&amp;M: 30% off w/ cartwheel + $5 Target GC for $16.78"></div><br /><div>Thumb Score: +25 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=d7ab36c8adb511e693311639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236807&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Target&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9328999" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Target</a> has <b>Great Deals</b> on <b>2x 24-Ct 3 oz Purina Fancy Feast Cat Food Variety Packs </b>when you follow the instructions below. <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks hjc985, Silvery79 &amp; Waynez [LIST][*]Online[LIST][*]Click 'Subscribe to save 5%' and change<b> Quantity to 2</b> (priced $11.99 or less)in cart (Mix-and-Match)</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/G5o2g_gJr1E" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 17:38:08 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>hjc985</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9328999-2x-24-cans-3-oz-purina-fancy-feast-variety-pack-5-target-gc-for-22-78-w-5-s-s-or-b-m-30-off-w-cartwheel-5-target-gc-for-16-78?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9328999-2x-24-cans-3-oz-purina-fancy-feast-variety-pack-5-target-gc-for-22-78-w-5-s-s-or-b-m-30-off-w-cartwheel-5-target-gc-for-16-78?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Couple Fluxx games on sale @ amazon $6.99 , $7.85</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/Ztwe9L75a6k/9359795-couple-fluxx-games-on-sale-amazon-6-99-7-85</link>
                                                <description><![CDATA[Frys.com has a couple of *Looney Labs Fluxx Card Games* on sale for the prices listed below. *Shipping is free* or select free store pickup where available. Thanks scottydog911[LIST][*]Looney Labs...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//1/2/2/6/9/8/5/200x200/4830011.thumb" alt="Couple Fluxx games on sale @ amazon $6.99 , $7.85"></div><br /><div>Thumb Score: +33 </div><div>Frys.com has a couple of <b>Looney Labs Fluxx Card Games</b> on sale for the prices listed below. <b><font color="#006400">Shipping is free</font></b> or select free store pickup where available. Thanks scottydog911[LIST][*]<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=ad35ded0adb411e6ba4efead90eb2890&amp;au=bc8574d68c2f11e29f3b6ee3110916cf&amp;pno=236811&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Looney+Labs+Fluxx+5+0+Card+Game&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9359795" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Looney Labs Fluxx 5.0 Card Game</a> <b>$6.99</b>[*]<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=ad35ded0adb411e6ba4efead90eb2890&amp;au=bc8574d68c2f11e29f3b6ee3110916cf&amp;pno=236811&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=Looney+Labs+Star+Fluxx+Card+Game&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9359795" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Looney Labs Star Fluxx Card Game</a> <b>$7.99</b>.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/Ztwe9L75a6k" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 17:30:16 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>scottydog911</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9359795-couple-fluxx-games-on-sale-amazon-6-99-7-85?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9359795-couple-fluxx-games-on-sale-amazon-6-99-7-85?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Philips Hue White & Color A19 Starter Pack (1st Gen, Refurbished)  $90 + Free Shipping]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/jOSbBAQS4jw/9325439-philips-hue-white-color-a19-starter-pack-1st-gen-refurbished-90-free-shipping</link>
                                                <description><![CDATA[Pie O My via Newegg has *Philips Hue White & Color A19 Starter Pack for**iOS and Android *(1st Gen, 426353, Refurbished) on sale for *$89.99*. *Shipping is free*. Thanks Cookie21 & TDMVP73

Note,...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//2/8/9/1/1/200x200/4829927.thumb" alt="Philips Hue White &amp; Color A19 Starter Pack (1st Gen, Refurbished)  $90 + Free Shipping"></div><br /><div>Thumb Score: +33 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=f7ed3424adb311e6a4971639df3e8759&amp;au=d469ad8c803211e2ad6402cf009d15c2&amp;pno=236815&amp;lno=1&amp;afsrc=1&amp;mon=0&amp;mid=701&amp;rnd=0&amp;mrnd=0&amp;subid=4623&amp;subrnd=0&amp;trd=Pie+O+My+via+Newegg&amp;ref=0" target="_blank" data-product-forum="Hot Deals" data-product-products="9325439" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow noreferrer">Pie O My via Newegg</a> has <b>Philips Hue White &amp; Color A19 Starter Pack for</b><b>iOS and Android </b>(1st Gen, 426353, Refurbished) on sale for <b>$89.99</b>. <b><font color="#006400">Shipping is free</font></b>. Thanks Cookie21 &amp; TDMVP73 <br />
<br />
Note, you also earn 10x eggpoints (up to 900 points) from the Pie O My via Newegg listing.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/jOSbBAQS4jw" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 17:24:26 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>Cookie21</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9325439-philips-hue-white-color-a19-starter-pack-1st-gen-refurbished-90-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9325439-philips-hue-white-color-a19-starter-pack-1st-gen-refurbished-90-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Board Games & More: Monopoly Here & Now, Scrabble Game, Operation, Twister Moves Hip Hop & More $6 Each+ Free In-Store Pickup via Toys R Us]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/RhxWBr-3dy4/9366619-board-games-more-monopoly-here-now-scrabble-game-operation-twister-moves-hip-hop-more-6-each-free-in-store-pickup-via-toys-r-us</link>
                                                <description><![CDATA[Toys R Us has several *Board Games* on Sale. Select in-store pick up or get *Free shipping* on orders $19 or more. Thanks Discombobulated
[LIST][*]Monopoly Here & Now Game *$6*[*]Monopoly SpongBob...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4829503.thumb" alt="Board Games &amp; More: Monopoly Here &amp; Now, Scrabble Game, Operation, Twister Moves Hip Hop &amp; More $6 Each+ Free In-Store Pickup via Toys R Us"></div><br /><div>Thumb Score: +62 </div><div>Toys R Us has several <b>Board Games</b> on Sale. Select in-store pick up or get <font color="DarkGreen"><b>Free shipping</b></font> on orders $19 or more. Thanks Discombobulated<br />
[LIST][*]<font color="Black"><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=6d9d7d38adae11e693311639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236803&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Monopoly+Here+amp+Now+Game&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366619" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Monopoly Here &amp; Now Game</a> <b>$6</b></font>[*]<font color="Black"><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=6d9d7d38adae11e693311639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236803&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=Monopoly+SpongBob+SquarePants+Ed&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366619" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Monopoly SpongBob SquarePants Edition</a> <b>$6</b></font></div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/RhxWBr-3dy4" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 16:45:29 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>Discombobulated</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366619-board-games-more-monopoly-here-now-scrabble-game-operation-twister-moves-hip-hop-more-6-each-free-in-store-pickup-via-toys-r-us?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366619-board-games-more-monopoly-here-now-scrabble-game-operation-twister-moves-hip-hop-more-6-each-free-in-store-pickup-via-toys-r-us?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Samsung Gear S2 Smartwatch w/ Heartrate Monitor  from $159 + Free Shipping</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/2_WZ0bnIs9s/9365291-samsung-gear-s2-smartwatch-w-heartrate-monitor-from-159-free-shipping</link>
                                                <description>BuyDig has a couple of styles of *Samsung Gear S2 Bluetooth/Wifi Smartwatch w/ Heartrate Monitor* on sale from as low as *$159*. Each one comes with a *bonus 2600mAh Keychain Power Bank* (added to...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4828691.thumb" alt="Samsung Gear S2 Smartwatch w/ Heartrate Monitor  from $159 + Free Shipping"></div><br /><div>Thumb Score: +28 </div><div>BuyDig has a couple of styles of <b>Samsung Gear S2 Bluetooth/Wifi Smartwatch w/ Heartrate Monitor</b> on sale from as low as <b>$159</b>. Each one comes with a <b>bonus 2600mAh Keychain Power Bank</b> (added to cart automatically). <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks viraj_parekh<br />
[LIST][*]<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=95e8b964ada511e6aa7b1639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236755&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Samsung+Gear+S2+Bluetooth+Wifi+S&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9365291" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Samsung Gear S2 Bluetooth/Wifi Smartwatch w/ HRM</a> (Dark Gray) <b>$159</b></div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/2_WZ0bnIs9s" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:37:35 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>viraj_parekh</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9365291-samsung-gear-s2-smartwatch-w-heartrate-monitor-from-159-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9365291-samsung-gear-s2-smartwatch-w-heartrate-monitor-from-159-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Nike+ Members Coupon: Men's, Women's & Children's' Clearance  25% Off + Free Shipping]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/m4Y02ZPX2Ec/9366403-nike-members-coupon-men-s-women-s-children-s-clearance-25-off-free-shipping</link>
                                                <description><![CDATA[Nike.com offers an *Extra 25% Off Men's, Women's & Children's Clearance Items* exclusively for Nike+ Members (free to join) when you apply promo code EARLY25 at checkout. *Shipping is free*. Thanks...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//2/5/1/6/8/8/6/200x200/4828543.thumb" alt="Nike+ Members Coupon: Men's, Women's &amp; Children's' Clearance  25% Off + Free Shipping"></div><br /><div>Thumb Score: +77 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=d3b3059eada811e698251639df3e8759&amp;au=7f81beb0e93a11e2816eea447bee2792&amp;pno=236783&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Nike+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366403" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Nike.com</a> offers an <b>Extra 25% Off Men's, Women's &amp; Children's Clearance Items</b> exclusively for Nike+ Members (<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=d3b3059eada811e698251639df3e8759&amp;au=7f81beb0e93a11e2816eea447bee2792&amp;pno=236783&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=free+to+join&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366403" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">free to join</a>) when you apply promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='997191'><strong>EARLY25</strong></span></span> at checkout. <b><font color="#006400">Shipping is free</font></b>. Thanks StrawMan86</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/m4Y02ZPX2Ec" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:20:28 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>StrawMan86</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366403-nike-members-coupon-men-s-women-s-children-s-clearance-25-off-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366403-nike-members-coupon-men-s-women-s-children-s-clearance-25-off-free-shipping?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>ecobee3 Smarter WiFi Thermostat w/ Remote Sensor (2nd Gen)  $199 + Free S/H</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/oUk4uPDqmNs/9366427-ecobee3-smarter-wifi-thermostat-w-remote-sensor-2nd-gen-199-free-s-h</link>
                                                <description>Home Depot.com also has *ecobee3 Smarter WiFi Thermostat w/ Remote Sensor* (2nd Gen) on sale for *$199*. *Shipping is free*. Thanks jmuiceman
Ecobee (direct) also their own *ecobee3 Smarter WiFi...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//2/8/9/1/1/200x200/4828535.thumb" alt="ecobee3 Smarter WiFi Thermostat w/ Remote Sensor (2nd Gen)  $199 + Free S/H"></div><br /><div>Thumb Score: +79 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=5e09e52aadc511e697151639df3e8759&amp;au=5e4c464c891911e2bcf4c65220ff958c&amp;pno=236799&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Home+Depot+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366427" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Home Depot.com</a> also has <b>ecobee3 Smarter WiFi Thermostat w/ Remote Sensor</b> (2nd Gen) on sale for <b>$199</b>. <b><font color="#006400">Shipping is free</font></b>. Thanks jmuiceman<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=5e09e52aadc511e697151639df3e8759&amp;au=5e4c464c891911e2bcf4c65220ff958c&amp;pno=236799&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366427" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow"><br /></a><br />
<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=5e09e52aadc511e697151639df3e8759&amp;au=5e4c464c891911e2bcf4c65220ff958c&amp;pno=236799&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366427" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Ecobee</a> (direct) also their own <b>ecobee3 Smarter WiFi Thermostat w/ Remote Sensor</b> (2nd Gen) on sale for <b>$199</b>. <b><font color="#006400">Shipping is free</font></b>.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/oUk4uPDqmNs" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:17:14 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>jmuiceman</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366427-ecobee3-smarter-wifi-thermostat-w-remote-sensor-2nd-gen-199-free-s-h?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366427-ecobee3-smarter-wifi-thermostat-w-remote-sensor-2nd-gen-199-free-s-h?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Rosewill 6' 3.5mm Flat Audio Cable  Free & More after Rebate + FS]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/IaIWSgrVl1w/9366355-rosewill-6-3-5mm-flat-audio-cable-free-more-after-rebate-fs</link>
                                                <description><![CDATA[Newegg has a few *Rosewill Products* for *Free *after Rebate. *Shipping is free*. Thanks TDMVP73
[LIST][*] Rosewill 1' Green Cat 6A Shielded Twisted Pair 550MHz Network Ethernet Cable...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4830751.thumb" alt="Rosewill 6' 3.5mm Flat Audio Cable  Free &amp; More after Rebate + FS"></div><br /><div>Thumb Score: +78 </div><div>Newegg has a few <b>Rosewill Products</b> for <b>Free </b>after Rebate. <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks TDMVP73<br />
[LIST][*] <a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=95589cfcadba11e6aa7b1639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236795&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Rosewill+1+Green+Cat+6A+Shielded&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366355" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Rosewill 1' Green Cat 6A Shielded Twisted Pair 550MHz Network Ethernet Cable</a> (RCNC-12025)[LIST][*]$4 - <a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=95589cfcadba11e6aa7b1639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236795&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=36+4+rebate&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366355" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">$4 rebate</a> =  <b>Free</b></div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/IaIWSgrVl1w" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:10:25 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>TDMVP73</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366355-rosewill-6-3-5mm-flat-audio-cable-free-more-after-rebate-fs?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366355-rosewill-6-3-5mm-flat-audio-cable-free-more-after-rebate-fs?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>FreedomPop Global 3-in-1 SIM Kit + Free 1GB Bonus - FreedomPop - $0.25</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/oHW7gvm7KWs/9363891-freedompop-global-3-in-1-sim-kit-free-1gb-bonus-freedompop-0-25</link>
                                                <description>FreedomPop.com is offering their *FreedomPop Global 3-In-1 SIM Kit + 1GB Bonus Data* on sale for *$0.25*. *Shipping is free*. Thanks Jbh98

Note, FreedomPop service is compatible with most unlocked...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//8/6/6/3/0/200x200/4828455.thumb" alt="FreedomPop Global 3-in-1 SIM Kit + Free 1GB Bonus - FreedomPop - $0.25"></div><br /><div>Thumb Score: +51 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=d89bb0d6ada011e693a11639df3e8759&amp;au=17dd8a58a31411e2999892d10e272585&amp;pno=236791&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=FreedomPop+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9363891" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">FreedomPop.com</a> is offering their <b>FreedomPop Global 3-In-1 SIM Kit + 1GB Bonus Data</b> on sale for <b>$0.25</b>. <b><font color="#006400">Shipping is free</font></b>. Thanks Jbh98<br />
<br />
Note, FreedomPop service is compatible with most unlocked GSM AT&amp;T or T-Mobile Phones (Android 4.0+ or iOS 7.0+)<br />
<br />
Includes[LIST][*]FreedomPop Global SIM Kit (Voice/Text/Data Bundle)</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/oHW7gvm7KWs" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:07:00 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>Jbh98</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9363891-freedompop-global-3-in-1-sim-kit-free-1gb-bonus-freedompop-0-25?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9363891-freedompop-global-3-in-1-sim-kit-free-1gb-bonus-freedompop-0-25?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Converse Chuck Taylor All Star Shoes: Hi Top $30, Low Top  $26 + Free S&H w/ Nike+ Acct]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/d7I7H9Qw4WU/9226847-converse-chuck-taylor-all-star-shoes-hi-top-30-low-top-26-free-s-h-w-nike-acct</link>
                                                <description><![CDATA[UPDATE: Price is now even lower!

Nike.com has select *Converse Shoes* on sale as listed below w/ an additional* 25% Off *when you sign in or join your Nike+ account [free to join] and enter promo...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//6/3/5/4/7/1/200x200/4828439.thumb" alt="Converse Chuck Taylor All Star Shoes: Hi Top $30, Low Top  $26 + Free S&amp;H w/ Nike+ Acct"></div><br /><div>Thumb Score: +189 </div><div>UPDATE: Price is now even lower!<br />
<br />
<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=9eeece5eada011e6bfc71639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236787&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Nike+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9226847" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Nike.com</a> has select <b>Converse Shoes</b> on sale as listed below w/ an additional<b> 25% Off </b>when you sign in or join your Nike+ account [<a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=9eeece5eada011e6bfc71639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236787&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=free+to+join&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9226847" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">free to join</a>] and enter promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='997191'><strong>EARLY25</strong></span></span> at checkout. <b><font color="#006400">Shipping is free</font></b> w/ a Nike+ account. Thanks couponmit</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/d7I7H9Qw4WU" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:06:32 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>couponmit</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9226847-converse-chuck-taylor-all-star-shoes-hi-top-30-low-top-26-free-s-h-w-nike-acct?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9226847-converse-chuck-taylor-all-star-shoes-hi-top-30-low-top-26-free-s-h-w-nike-acct?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[Battlefield 1 & Titanfall 2 Deluxe Digital Bundle for $75 at Microsoft.com (Xbox One)]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/Hkqqi-ITWGc/9366311-battlefield-1-titanfall-2-deluxe-digital-bundle-for-75-at-microsoft-com-xbox-one</link>
                                                <description>Microsoft Store.com has *Battlefield 1 + Titanfall 2 Deluxe Edition* (Xbox One Digital Download) on sale for *$75* valid for *Xbox Live Gold Members* only. Thanks matthewhsh

Note, must be logged...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//8/6/6/3/0/200x200/4828443.thumb" alt="Battlefield 1 &amp; Titanfall 2 Deluxe Digital Bundle for $75 at Microsoft.com (Xbox One)"></div><br /><div>Thumb Score: +51 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=a8cff948ada011e6bfc71639df3e8759&amp;au=17dd8a58a31411e2999892d10e272585&amp;pno=236779&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Microsoft+Store+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366311" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Microsoft Store.com</a> has <b>Battlefield 1 + Titanfall 2 Deluxe Edition</b> (Xbox One Digital Download) on sale for <b>$75</b> valid for <b>Xbox Live Gold Members</b> only. Thanks matthewhsh<br />
<br />
Note, must be logged into your account w/ an active Xbox Live Gold Membership to view sale price.<br />
<br />
Includes[LIST][*]Battlefield 1 (Xbox One Digital Game)</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/Hkqqi-ITWGc" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 15:05:00 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>matthewhsh</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366311-battlefield-1-titanfall-2-deluxe-digital-bundle-for-75-at-microsoft-com-xbox-one?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366311-battlefield-1-titanfall-2-deluxe-digital-bundle-for-75-at-microsoft-com-xbox-one?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Just Cause 3 (PC) $9.69 or less @ cdkeys.com</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/TUtEt0fnk6M/9316735-just-cause-3-pc-9-69-or-less-cdkeys-com</link>
                                                <description><![CDATA[CDKeys.com has *Just Cause 3* (PC Digital Download) on sale for *$9.69. *Thanks jvkiller28

Note, you may also click here and 'Like' CDKeys on Facebook to receive an additional 5% off unique...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//6/3/5/4/7/1/200x200/4828211.thumb" alt="Just Cause 3 (PC) $9.69 or less @ cdkeys.com"></div><br /><div>Thumb Score: +47 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=f9d992d0ad9c11e6a2b11639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236775&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=CDKeys+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9316735" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">CDKeys.com</a> has <b>Just Cause 3</b> (PC Digital Download) on sale for <b>$9.69. </b>Thanks jvkiller28<br />
<br />
Note, you may also <a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=f9d992d0ad9c11e6a2b11639df3e8759&amp;au=be91cb1a7a2a11e2a40bb6794f5fda6a&amp;pno=236775&amp;lno=2&amp;afsrc=1&amp;mon=1&amp;trd=click+here&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9316735" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">click here</a> and 'Like' CDKeys on Facebook to receive an additional 5% off unique discount code (apply at checkout). Must be logged in to CDKeys for code to apply</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/TUtEt0fnk6M" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 14:40:09 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>jvkiller28</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9316735-just-cause-3-pc-9-69-or-less-cdkeys-com?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9316735-just-cause-3-pc-9-69-or-less-cdkeys-com?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>2TB WD My Passport Portable External Hard Drive + $3.50 in Rakuten Super Points $70 w/ VISA Checkout + Free SH</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/1mImRYZjH7k/9365951-2tb-wd-my-passport-portable-external-hard-drive-3-50-in-rakuten-super-points-70-w-visa-checkout-free-sh</link>
                                                <description>Dell via Rakuten has *2TB WD My Passport Portable External Hard Drive *(WDBYFT0020BBK-WESN) for $89.99 - $20 w/ promo code VISAFRIDAY (must use VISA checkout) = *$69.99*. *Shipping is free*. Thanks...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4828175.thumb" alt="2TB WD My Passport Portable External Hard Drive + $3.50 in Rakuten Super Points $70 w/ VISA Checkout + Free SH"></div><br /><div>Thumb Score: +25 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=92ee16faad9b11e693311639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236767&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Dell+via+Rakuten&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9365951" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Dell via Rakuten</a> has <b>2TB WD My Passport Portable External Hard Drive </b>(WDBYFT0020BBK-WESN) for $89.99 - $20 w/ promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='997179'><strong>VISAFRIDAY</strong></span></span> (must use VISA checkout) = <b>$69.99</b>. <font color="DarkGreen"><b>Shipping is free</b></font>. Thanks DJ3xclusive<br />
<br />
*You may also earn <b>5% back in Rakuten Points</b> (350 Points = <b>$3.50</b>) w/ promo code <span data-role='couponCode'><span class='icon icon-coupon inlineCouponIcon'></span><span class='code' data-couponid='982695'><strong>REWARDME</strong></span></span>. You must be logged into your Rakuten Account in order to apply)</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/1mImRYZjH7k" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 14:30:32 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>DJ3xclusive</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9365951-2tb-wd-my-passport-portable-external-hard-drive-3-50-in-rakuten-super-points-70-w-visa-checkout-free-sh?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9365951-2tb-wd-my-passport-portable-external-hard-drive-3-50-in-rakuten-super-points-70-w-visa-checkout-free-sh?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>GoPro Hero4 Silver + Extra battery + Dual battery charger + SanDisk 16GB MicroSD $200 @ Costco.com</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/rc-5ibf78a8/9366359-gopro-hero4-silver-extra-battery-dual-battery-charger-sandisk-16gb-microsd-200-costco-com</link>
                                                <description>Costco Wholesale.com has *GoPro HERO4 1080p Action Camera Bundle* on sale for *$199.99* valid for *Costco Members* only. *Shipping is free*. Thanks TeddyPointO
Note, must login to account to view...</description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//8/6/6/3/0/200x200/4828123.thumb" alt="GoPro Hero4 Silver + Extra battery + Dual battery charger + SanDisk 16GB MicroSD $200 @ Costco.com"></div><br /><div>Thumb Score: +87 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=d24316ecad9711e69c231639df3e8759&amp;au=17dd8a58a31411e2999892d10e272585&amp;pno=236771&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Costco+Wholesale+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9366359" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Costco Wholesale.com</a> has <b>GoPro HERO4 1080p Action Camera Bundle</b> on sale for <b>$199.99</b> valid for <b>Costco Members</b> only. <font color="#006400"><b>Shipping is free</b></font>. Thanks TeddyPointO<br />
Note, must login to account to view sale price.<br />
<br />
Includes[LIST][*]GoPro HERO4 1080p Action Camera (Silver)[*]GoPro Battery[*]GoPro Dual Battery Charge</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/rc-5ibf78a8" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 14:03:00 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>TeddyPointO</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366359-gopro-hero4-silver-extra-battery-dual-battery-charger-sandisk-16gb-microsd-200-costco-com?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366359-gopro-hero4-silver-extra-battery-dual-battery-charger-sandisk-16gb-microsd-200-costco-com?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title><![CDATA[70" lg 70uh6350 4k uhd hdtv Costco $999]]></title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/wItWJoQqiJw/9366363-70-lg-70uh6350-4k-uhd-hdtv-costco-999</link>
                                                <description><![CDATA[Costco Wholesale.com has *70" LG 70UH6350 4K UHD HDR Smart LED HDTV w/ webOS 3.0* on sale for *$999.99* valid for *Costco Members* only. *Shipping is free*. Thanks pcpcpc

Note, must login to...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//8/6/6/3/0/200x200/4828079.thumb" alt="70&quot; lg 70uh6350 4k uhd hdtv Costco $999"></div><br /><div>Thumb Score: +115 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=41a007c8ad9511e698251639df3e8759&amp;au=17dd8a58a31411e2999892d10e272585&amp;pno=236763&amp;lno=1&amp;afsrc=1&amp;mon=1&amp;trd=Costco+Wholesale+com&amp;ref=1" target="_blank" data-product-forum="Deal Talk" data-product-products="9366363" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Costco Wholesale.com</a> has <b>70&quot; LG 70UH6350 4K UHD HDR Smart LED HDTV w/ webOS 3.0</b> on sale for <b>$999.99</b> valid for <b>Costco Members</b> only. <b><font color="#006400">Shipping is free</font></b>. Thanks pcpcpc<br />
<br />
Note, must login to account to view sale price.<br />
<br />
Specs[LIST][*]Resolution: 3840x2160[*]Refresh Rate: 240Hz (TruMotion)[*]Inputs:</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/wItWJoQqiJw" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 13:45:00 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>pcpcpc</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9366363-70-lg-70uh6350-4k-uhd-hdtv-costco-999?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9366363-70-lg-70uh6350-4k-uhd-hdtv-costco-999?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
          <item>
                                            <title>Grand Kingdom PS4 $20.99 Amazon</title>
                                                <link>http://feedproxy.google.com/~r/SlickdealsnetFP/~3/zFUjt4PmedU/9354159-grand-kingdom-ps4-20-99-amazon</link>
                                                <description><![CDATA[Amazon.com (for Amazon *Prime members* only), has *Grand Kingdom: Launch Day Edition* (PS4) on sale for *$20.99*. *Shipping is free* with your Prime. Thanks Prideless

*Deal Editor's Notes & Price...]]></description>
                                                <content:encoded><![CDATA[<div><img src="https://static.slickdealscdn.com/attachment//4/7/6/0/3/200x200/4828055.thumb" alt="Grand Kingdom PS4 $20.99 Amazon"></div><br /><div>Thumb Score: +35 </div><div><a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=9daec33aad9311e69c231639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236743&amp;lno=1&amp;sdfib=1&amp;afsrc=1&amp;mon=1&amp;trd=Amazon+com&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9354159" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">Amazon.com</a> (for Amazon <b>Prime members</b> only), has <b>Grand Kingdom: Launch Day Edition</b> (PS4) on sale for <b>$20.99</b>. <font color="DarkGreen"><b>Shipping is free</b></font> with your Prime. Thanks Prideless<br />
<br />
<b>Deal Editor's Notes &amp; Price Research:</b> [LIST][*]Don't have Amazon Prime?[LIST][*]Students can get a <a href="https://slickdeals.net/?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2&amp;pv=9daec33aad9311e69c231639df3e8759&amp;au=f13af86688ad11e296b08a8e19448ad2&amp;pno=236743&amp;lno=2&amp;sdfib=1&amp;afsrc=1&amp;mon=1&amp;trd=free+6+Month+Amazon+Prime+trial&amp;ref=1" target="_blank" data-product-forum="Hot Deals" data-product-products="9354159" data-outclick-typeofoutclick="Post Content Prebuilt Link" data-product-exitWebsite="slickdeals.net" rel="nofollow">free 6-Month Amazon Prime trial</a> with free 2-day shipping, unlimited music, unlimited video streaming &amp; more.</div>
<img src="http://feeds.feedburner.com/~r/SlickdealsnetFP/~4/zFUjt4PmedU" height="1" width="1" alt=""/>]]></content:encoded>
                                                <pubDate>Fri, 18 Nov 2016 13:33:28 GMT</pubDate>
                                                <category domain="https://slickdeals.net/">Frontpage Deals</category>
                                                <dc:creator>Prideless</dc:creator>
                                                <guid isPermaLink="false">https://slickdeals.net/f/9354159-grand-kingdom-ps4-20-99-amazon?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</guid>
                                    <feedburner:origLink>https://slickdeals.net/f/9354159-grand-kingdom-ps4-20-99-amazon?utm_source=rss&amp;utm_content=fp&amp;utm_medium=RSS2</feedburner:origLink></item>
      </channel>
</rss>`
*/
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
