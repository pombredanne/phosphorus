package app

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/crowdmob/goamz/dynamodb"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"willstclair.com/phosphorus/db"
	"willstclair.com/phosphorus/id"
)

type Account struct {
	Username string `dynamodb:"_hash"`
	Salt     string `dynamodb:"salt"`
	Password string `dynamodb:"password"`
}

func kdf(salt, password []byte) []byte {
	hash, err := scrypt.Key(password, salt, 1<<14, 8, 1, 32)
	if err != nil {
		panic(err)
	}
	return hash
}

func newSalt() []byte {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	return salt
}

func b2s(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func s2b(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func (a *Account) Authenticate(password string) (err error) {

	p := []byte(password)
	hash := kdf(s2b(a.Salt), p)
	if subtle.ConstantTimeCompare(s2b(a.Password), hash) != 1 {
		err = fmt.Errorf("bad password")
	}
	return
}

type Session struct {
	Id       int64  `dynamodb:"_hash"`
	Username string `dynamodb:"username"`
	Expires  int64  `dynamodb:"expires"`
}

func NewSession(username string, idGen *id.Generator) *Session {
	expires := time.Now().UTC().Add(12 * time.Hour)
	return &Session{
		Id:       idGen.SafeId(),
		Username: username,
		Expires:  expires.Unix()}
}

func (s *Session) Cookie() *http.Cookie {
	return &http.Cookie{
		Name:    "phosphorus",
		Value:   strconv.FormatInt(s.Id, 10),
		Expires: time.Unix(s.Expires, 0)}
}

func (s *Session) Valid() bool {
	return s.Expires > time.Now().UTC().Unix()
}

func CreateAccountHandler(tpl map[string]*template.Template, accountTbl *dynamodb.Table, sessionTbl *dynamodb.Table, idGen *id.Generator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("enrolling")
		switch r.Method {
		case "GET":
			log.Println("GET")
			tpl["enroll.html"].ExecuteTemplate(w, "layout", nil)
			return
		case "POST":
			break
		default:
			w.WriteHeader(405)
			return
		}

		if r.FormValue("password") != r.FormValue("confirm") {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Password mismatch")
			return
		}

		username := strings.ToLower(r.FormValue("username"))
		matched, _ := regexp.MatchString("^[a-z][a-z0-9]{2,15}$", username)
		if !matched {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid username")
			return
		}

		salt := newSalt()
		password := []byte(r.FormValue("password"))
		a := &Account{
			Username: r.FormValue("username"),
			Salt:     b2s(salt),
			Password: b2s(kdf(salt, password))}

		err := db.CreateItem(accountTbl, a)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error creating account")
			return
		}

		cookieAndGo(w, username, sessionTbl, idGen)
	}
}

func LoginHandler(tpl map[string]*template.Template, accountTbl *dynamodb.Table, sessionTbl *dynamodb.Table, idGen *id.Generator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			tpl["login.html"].ExecuteTemplate(w, "layout", nil)
			return
		case "POST":
			break
		default:
			w.WriteHeader(405)
			return
		}

		acct := &Account{Username: r.FormValue("username")}
		err := db.GetItem(accountTbl, acct)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Bad authentication")
			return
		}

		err = acct.Authenticate(r.FormValue("password"))
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Bad authentication")
			return
		}

		cookieAndGo(w, acct.Username, sessionTbl, idGen)
	}
}

func DashboardHandler(tpl map[string]*template.Template,
	accountTbl *dynamodb.Table, sessionTbl *dynamodb.Table) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}

		cookie, err := r.Cookie("phosphorus")
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "No token")
			return
		}

		sessionId, err := strconv.ParseInt(cookie.Value, 10, 64)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad token")
			return
		}

		session := &Session{Id: sessionId}
		err = db.GetItem(sessionTbl, session)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Token not found")
			return
		}

		if !session.Valid() {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Old token")
			return
		}

		account := &Account{Username: session.Username}
		err = db.GetItem(accountTbl, account)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Account unavailable")
			return
		}

		// hey we're ok
		tpl["dashboard.html"].ExecuteTemplate(w, "layout", account)
	}
}

func cookieAndGo(w http.ResponseWriter, username string, sessionTbl *dynamodb.Table, idGen *id.Generator) {
	session := NewSession(username, idGen)
	err := db.CreateItem(sessionTbl, session)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error creating session")
		return
	}

	w.Header().Add("Set-Cookie", session.Cookie().String())
	w.Header().Add("Location", "/u/"+username)
	w.WriteHeader(303)
}
