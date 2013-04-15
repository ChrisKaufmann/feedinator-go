package main

import (
	"net/http"
	"database/sql"
	"strings"
	"strconv"
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
func sth(db *sql.DB,s string) *sql.Stmt {
	a, err := db.Prepare(s)
	if err != nil {
		print (s)
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
func unescape(s string) string {
	s = strings.Replace(s, "&#34;", "\"", -1)
	s = strings.Replace(s, "&#47;", "/", -1)
	return s
}

func toint(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}