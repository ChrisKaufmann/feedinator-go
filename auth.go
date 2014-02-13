package main

import (
	"fmt"
	"code.google.com/p/goauth2/oauth"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"github.com/msbranco/goconfig"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var oauthCfg = &oauth.Config{
	AuthURL:     "https://accounts.google.com/o/oauth2/auth",
	TokenURL:    "https://accounts.google.com/o/oauth2/token",
	Scope:       "https://www.googleapis.com/auth/userinfo.email",
	TokenCache:  oauth.CacheFile(cachefile),
}

func init() {

	c, err := goconfig.ReadConfigFile("config")
	if err != nil {
		panic(err)
	}
	oauthCfg.ClientId, err = c.GetString("Google", "ClientId")
	if err != nil {
		panic(err)
	}
	oauthCfg.ClientSecret, err = c.GetString("Google", "ClientSecret")
	if err != nil {
		panic(err)
	}
	url, err := c.GetString("Web","url")
	if err != nil {
		panic(err)
	}
	oauthCfg.RedirectURL = url+"oauth2callback"
}

// Start the authorization process
func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	print("In handleauth\n")
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	//redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

// Function that handles the callback from the Google server
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	print("code="+code)
	t := &oauth.Transport{oauth.Config: oauthCfg}

	// Exchange the received code for a token
	tok, err := oauthCfg.TokenCache.Token()
	if err != nil {
		print(err)
		tok, err = t.Exchange(code)
		if err != nil {
			print(err)
			panic(err.Error())
		}
		fmt.Printf("token cached in %v\n",oauthCfg.TokenCache)
	}
	print(tok)


	// Make the request.
	req, err := t.Client().Get(profileInfoURL)
	if err != nil {
		print(err)
		print("\n")
		panic(err.Error())
		return
	}
	defer req.Body.Close()
	body, _ := ioutil.ReadAll(req.Body)
	log.Println(string(body))
	//print(string(body))
	//body.id is the google id to use
	//set a cookie with the id, and random hash. then save the id/hash pair to db for lookup
	var f interface{}
	err = json.Unmarshal(body, &f)
	m := f.(map[string]interface{})
	print(m["email"].(string))
	if err != nil {
		panic(err.Error())
	}
	var authString = randomString(64)
	makeSureUserExists(m["email"].(string))
	_, err = stmtCookieIns.Exec(m["email"], hash(authString))

	if err != nil {
		panic(err.Error())
	}
	//set the cookie
	expire := time.Now().AddDate(1, 0, 0) // year expirey seems reasonable
	cookie := http.Cookie{Name: cookieName, Value: authString, Expires: expire}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/main", http.StatusFound)
}
func makeSureUserExists(e string) {
	var userId string
	err := stmtGetUsername.QueryRow(e).Scan(&userId)
	if err != nil {
		// in this case, emans there's no username - create one
		_, err = stmtInsertUser.Exec(e,e)
	}
}
func loggedIn(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		//just means that the cookie doesn't exist or we couldn't read it
		return false
	}
	tok := cookie.Value
	tokHash := hash(tok)
	var userId string
	err = stmtGet.QueryRow(tokHash).Scan(&userId) // WHERE number = 13
	if err != nil {
		return false //probably no rows in result set
	}

	if userId != "" {
		userName = userId
		return true
	} else {
		return false
	}
	return false
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
