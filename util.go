package main

import (
	"crypto/sha1"
	"database/sql"
	"time"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
func shufflei(slice []int) []int {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}
func shuffleFeeds(slice []Feed) []Feed {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
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
	var codes = map[string]string{
		"&amp;":               "&",
		"&nbsp;":              " ",
		"&acirc;&#128;&#153;": "'",
	}
	for k, v := range codes {
		s = strings.Replace(s, k, v, -1)
	}
	return s
}
func randomString(l int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
func hash(s string) string {
	h := sha512.New()
	h.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
