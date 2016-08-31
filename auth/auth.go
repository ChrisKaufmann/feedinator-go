package auth

import (
	"code.google.com/p/goauth2/oauth"
	"database/sql"
	"encoding/json"
	"flag"
	u "github.com/ChrisKaufmann/goutils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/msbranco/goconfig"
	"io/ioutil"
	"net/http"
	"time"
)

var oauthCfg = &oauth.Config{
	AuthURL:    "https://accounts.google.com/o/oauth2/auth",
	TokenURL:   "https://accounts.google.com/o/oauth2/token",
	Scope:      "https://www.googleapis.com/auth/userinfo.email",
	TokenCache: oauth.CacheFile(cachefile),
}

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo"
const cachefile = "/dev/null"

var (
	MyURL                string
	db                   *sql.DB
	cookieName           string = "auth"
	environment          string = "production"
	stmtCookieIns        *sql.Stmt
	stmtGetUserID        *sql.Stmt
	stmtInsertUser       *sql.Stmt
	stmtGetUserBySession *sql.Stmt
	stmtSessionExists    *sql.Stmt
	stmtLogoutSession    *sql.Stmt
	stmtGetAllUsers      *sql.Stmt
)

func CookieName(c string) {
	cookieName = c
}
func Environment(e string) {
	environment = e
}
func DB(d *sql.DB) {
	db = d
	var err error
	stmtCookieIns, err = u.Sth(db, "INSERT INTO sessions (user_id,session_hash) VALUES( ? ,?  )")
	if err != nil {
		glog.Fatalf(" DB(): u.sth(stmtCookieIns) %s", err)
	}
	stmtGetUserID, err = u.Sth(db, "select id from users where email = ?")
	if err != nil {
		glog.Fatalf(" DB(): u.sth(stmtGetUserID) %s", err)
	}
	stmtInsertUser, err = u.Sth(db, "insert into users (id,email) values (?,?) ")
	if err != nil {
		glog.Fatalf(" DB(): u.sth(stmtInsertUser) %s", err)
	}
	stmtGetUserBySession, err = u.Sth(db, "select users.id, users.email from users, sessions where users.id=sessions.user_id and sessions.session_hash=?")
	if err != nil {
		glog.Fatalf(" DB(): u.sth(stmtGetUserBySession) %s", err)
	}
	stmtSessionExists, err = u.Sth(db, "select user_id from sessions where session_hash=?")
	if err != nil {
		glog.Fatalf(" DB(): u.sth(stmtSessionExists) %s", err)
	}
	stmtLogoutSession, err = u.Sth(db, "delete from sessions where session_hash=? limit 1")
	if err != nil {
		glog.Fatalf(" DB(): u.sth(stmtLogoutSession) %s", err)
	}
	stmtGetAllUsers, err = u.Sth(db, "select id, email from users where 1")
	if err != nil {
		glog.Fatalf("sth(db, select id, email from users where 1): %s", err)
	}
}
func init() {
	flag.Parse()
	c, err := goconfig.ReadConfigFile("config")
	if err != nil {
		glog.Warningf("init(): readconfigfile(config)")
		oauthCfg.ClientSecret = "a"
		oauthCfg.ClientId = "b"
		MyURL = "c"
		oauthCfg.RedirectURL = "d"
		return
	}
	oauthCfg.ClientId, err = c.GetString("Google", "ClientId")
	if err != nil {
		glog.Fatal("init(): readconfigfile(Google.ClientId), using dummy")
	}
	oauthCfg.ClientSecret, err = c.GetString("Google", "ClientSecret")
	if err != nil {
		glog.Fatal("init(): readconfigfile(Google.ClientSecret)")
	}
	url, err := c.GetString("Web", "url")
	MyURL = url
	if err != nil {
		glog.Fatal("init(): readconfigfile(Web.url)")
	}
	oauthCfg.RedirectURL = url + "/oauth2callback"
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		//just means that the cookie doesn't exist or we couldn't read it
		glog.Infof("HandleLogout: No cookie to logut %s", err)
		return
	}
	tokHash := cookie.Value
	if !SessionExists(tokHash) {
		glog.Info("HandleLogout: No matching sessions")
	}
	_, err = stmtLogoutSession.Exec(tokHash)
	if err != nil {
		glog.Errorf("HandleLougout: stmtLogoutSession.Exec(%s): %s", tokHash, err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Start the authorization process
func HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

//simulate a demo login, create the cookie, make sure the demo user exists, create the session
func DemoUser(w http.ResponseWriter, r *http.Request) {
	demo_email := "chriskaufmann@gmail.com"
	var us User
	var err error
	if !UserExists(demo_email) {
		us, err = AddUser(demo_email)
		if err != nil {
			glog.Errorf("DemoUser(w,r)AddUser(%s): %s", demo_email, err)
			return
		}
	} else {
		us, err = GetUserByEmail(demo_email)
		if err != nil {
			glog.Errorf("DemoUser(w,r)GetUserByEmail(%s): %s", demo_email, err)
			return
		}
	}
	var authString = u.RandomString(64)
	//set the cookie
	err = us.AddSession(authString)
	if err != nil {
		glog.Errorf("DemoUser(w,r)AddUser(%s): %s", authString, err)
		return
	}
	expire := time.Now().AddDate(1, 0, 0) // year expirey seems reasonable
	cookie := http.Cookie{Name: cookieName, Value: authString, Expires: expire}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/main", http.StatusFound)
}

// Function that handles the callback from the Google server
func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{Config: oauthCfg}

	// Exchange the received code for a token
	_, err := oauthCfg.TokenCache.Token()
	if err != nil {
		_, err := t.Exchange(code)
		if err != nil {
			glog.Errorf("HandleOauth2Callback:oauthCfg.TokenCache.Token():t.Exchange(%s): %s", code, err)
		}
	}

	// Make the request.
	req, err := t.Client().Get(profileInfoURL)
	if err != nil {
		glog.Errorf("HandleOauth2Callback:t.Client().Get(%s): %s", profileInfoURL, err)
		return
	}
	defer req.Body.Close()
	body, _ := ioutil.ReadAll(req.Body)
	//body.id is the google id to use
	//set a cookie with the id, and random hash. then save the id/hash pair to db for lookup
	var f interface{}
	err = json.Unmarshal(body, &f)
	if err != nil {
		glog.Errorf("HandleOauth2Callback:json.Unmarshal(%s): %s", body, err)
		return
	}
	m := f.(map[string]interface{})
	var authString = u.RandomString(64)
	email := m["email"].(string)
	var us User
	if !UserExists(email) {
		glog.Infof("HandleOauth2Callback: creating new user %s", email)
		us, err = AddUser(email)
		if err != nil {
			glog.Errorf("HandleOauth2Callback:UserExists()AddUser(%s): %s", email, err)
		}
	} else {
		us, err = GetUserByEmail(email)
		if err != nil {
			glog.Errorf("HandleOauth2Callback:UserExists()GetUserEmail(%s): %s", email, err)
		}
	}

	err = us.AddSession(authString)

	if err != nil {
		glog.Errorf("HandleOauth2Callback:stmtCookieIns.Exec(%s,%s): %s", us.ID, authString, err)
	}
	//set the cookie
	expire := time.Now().AddDate(1, 0, 0) // year expirey seems reasonable
	cookie := http.Cookie{Name: cookieName, Value: authString, Expires: expire}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/main", http.StatusFound)
}
func LoggedIn(w http.ResponseWriter, r *http.Request) (bool, string) {
	if environment == "test" {
		return true, "chris"
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		//just means that the cookie doesn't exist or we couldn't read it
		glog.Info("LoggedIn(): No cookie")
		return false, ""
	}
	tokHash := cookie.Value
	if !SessionExists(tokHash) {
		return false, ""
	}
	us, err := GetUserBySession(tokHash)
	if err != nil {
		glog.Errorf("LoggedIn():GetUserBySession(%s): %s", tokHash, err)
		return false, ""
	}
	if us.ID != "" {
		return true, us.ID
	}
	return false, ""
}
