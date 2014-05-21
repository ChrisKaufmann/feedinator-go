package main

import (
	"database/sql"
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
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
func tostr(i int) string {
	s := strconv.Itoa(i)
	return s
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
	err = mc.Set(&memcache.Item{Key: key, Value: []byte(b), Expiration: t})
	return err
}
