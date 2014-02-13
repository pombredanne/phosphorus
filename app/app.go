package app

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/crowdmob/goamz/dynamodb"
	"github.com/crowdmob/goamz/s3"
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
	id := idGen.SafeId()
	log.Println("new session id %d", id)
	return &Session{
		Id:       id,
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
	accountTbl *dynamodb.Table, sessionTbl *dynamodb.Table, idGen *id.Generator, bucket *s3.Bucket) func(http.ResponseWriter, *http.Request) {
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

		if r.URL.Path[len(fmt.Sprintf("/u/%s", account.Username)):] == "/source" {
			handleSource(w, r, tpl, account, bucket)
			return
		}

		listResp, err := bucket.List(account.Username+"/", "", "", 0)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Problem with S3")
			return
		}
		expires := time.Now().UTC().Add(5 * time.Minute)
		policy := fmt.Sprintf(s3policy, account.Username,
			expires.Format("2006-01-02T15:04:05.000Z"))
		log.Println(policy)
		encodedPolicy := base64.StdEncoding.EncodeToString([]byte(policy))

		mac := hmac.New(sha1.New, []byte(sessionTbl.Server.Auth.SecretKey))
		mac.Write([]byte(encodedPolicy))
		sig := mac.Sum(nil)

		// hey we're ok
		tpl["dashboard.html"].ExecuteTemplate(w, "layout", &_hm{
			Keys:        listResp.Contents,
			Filename:    fmt.Sprintf("%s/%d", account.Username, idGen.SafeId()),
			Username:    account.Username,
			Policy:      encodedPolicy,
			Signature:   base64.StdEncoding.EncodeToString(sig),
			AccessKeyId: sessionTbl.Server.Auth.AccessKey})
	}
}

type _hm struct {
	Username    string
	Policy      string
	Signature   string
	AccessKeyId string
	Filename    string
	Keys        []s3.Key
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

var s3policy = `{"conditions":[["starts-with","$key","%s/"],{"bucket":"phosphorus-upload"},{"acl":"private"},["starts-with","$Content-Type","text/csv"],{"success_action_status":"200"}],"expiration":"%s"}`

func handleSource(w http.ResponseWriter, r *http.Request, tpl map[string]*template.Template, account *Account, bucket *s3.Bucket) {
	rc, err := bucket.GetReader(r.URL.Query().Get("key"))
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Problem opening file")
		return
	}
	defer rc.Close()
	rdr := csv.NewReader(rc)

	records := make([][]string, 0, 10)

	for i := 0; i < 10; i++ {
		record, err := rdr.Read()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Problem reading file")
			return
		}
		records = append(records, record)
	}

	tpl["source.html"].ExecuteTemplate(w, "layout", &_sourceDef{
		Rows:    records,
		Columns: rdr.FieldsPerRecord,
	})
}

type _sourceDef struct {
	Columns int
	Rows    [][]string
}

func UploadTemplateHandler(accountTbl *dynamodb.Table, sessionTbl *dynamodb.Table, idGen *id.Generator) func(http.ResponseWriter, *http.Request) {
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

		expires := time.Now().UTC().Add(1 * time.Minute)
		policy := fmt.Sprintf(s3policy, account.Username,
			expires.Format("2006-01-02T15:04:05.000Z"))
		log.Println(policy)
		encodedPolicy := base64.StdEncoding.EncodeToString([]byte(policy))

		mac := hmac.New(sha1.New, []byte(sessionTbl.Server.Auth.SecretKey))
		mac.Write([]byte(encodedPolicy))
		sig := mac.Sum(nil)

		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(&_uploadForm{
			Key:            fmt.Sprintf("%s/%d", account.Username, idGen.SafeId()),
			AWSAccessKeyId: sessionTbl.Server.Auth.AccessKey,
			Policy:         encodedPolicy,
			Signature:      base64.StdEncoding.EncodeToString(sig)})
	}
}

type _uploadForm struct {
	Key            string `json:"key"`
	AWSAccessKeyId string `json:"AWSAccessKeyId"`
	Policy         string `json:"policy"`
	Signature      string `json:"signature"`
}
