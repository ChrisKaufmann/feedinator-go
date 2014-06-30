package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var (
	mc = memcache.New("127.0.0.1:11211")
)

//puts path vars right into variables passed as params, until it runs out
//ex: pathVars(r,"/entry/",&id,&todo,&val) // populates id, todo, and val
func pathVars(r *http.Request, root string, vals ...*string) {
	a := strings.Split(r.URL.Path[len(root):], "/")
	for i := range vals {
		if len(a) > i {
			*vals[i] = a[i]
		} else {
			*vals[i] = ""
		}
	}
}
func sth(db *sql.DB, s string) *sql.Stmt {
	a, err := db.Prepare(s)
	if err != nil {
		print(s)
		panic(err)
	}
	return a
}

func evenodd(i int) string {
	if i%2 == 0 {
		return "even"
	}
	return "odd"
}
func tostr(i interface{}) string {
	switch t := i.(type) {
	case string:
		return i.(string)
	case int:
		return strconv.Itoa(i.(int))
	case bool:
		return strconv.FormatBool(i.(bool))
	case template.HTML:
		return string(i.(template.HTML))
	default:
		t = t
	}
	return i.(string)
}
func toint(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
func mcset(key string, i interface{}) (err error) {
	var timeout int32 = 86400
	err = mcsettime(key, i, timeout)
	return err
}
func mcsettime(key string, i interface{}, t int32) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	err = mc.Set(&memcache.Item{Key: MyURL + key, Value: []byte(b), Expiration: t})
	return err
}
func mcget(key string, i interface{}) error {
	thing, err := mc.Get(MyURL + key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(thing.Value, &i)
	return err
}
func mcdel(key string) (err error) {
	err = mc.Delete(MyURL + key)
	return err
}
func getHash(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%s", base64.URLEncoding.EncodeToString(h.Sum(nil)))
}
func escape_guid(s string) string {
	var htmlCodes = map[string]string{
		"&#34;":   "\"",
		"&#47;":   "/",
		"&#39;":   "'",
		"&#42;":   "*",
		"&#63;":   "?",
		"&#160;":  " ",
		"&#8216;": "'",
		"&#8220;": "'",
		"&#8221;": "'",
		"&#8211;": "-",
		"&#8230;": "...",
		"&#8594;": "->",
		"&quot;":  "'",
		"&amp;":   "&",
		"&#37;":   "%",
	}
	for k, v := range htmlCodes {
		s = strings.Replace(s, k, v, -1)
	}
	return s
}

func unescape(s string) string {
	s = strings.Replace(s, "&amp;#", "&#", -1)
	return s
}
