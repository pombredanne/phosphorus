package app

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"willstclair.com/phosphorus/db"
)

// models

type Account struct {
	Username string `dynamodb:"_hash"`
	Salt     string `dynamodb:"salt"`
	Password string `dynamodb:"password"`
}

func (a *Account) Authenticate(password string) (err error) {
	p := []byte(password)
	hash := kdf(s2b(a.Salt), p)
	if subtle.ConstantTimeCompare(s2b(a.Password), hash) != 1 {
		err = fmt.Errorf("bad password")
	}
	return
}

var hmacKey []byte

func init() {
	hmacKey, _ = scrypt.Key([]byte("very secure"), []byte{}, 1<<18, 8, 1, 16)
}

type Session struct {
	Id       int64  `dynamodb:"_hash"`
	Username string `dynamodb:"username"`
	Expires  int64  `dynamodb:"expires"`
}

func NewSession(username string, id int64) *Session {
	expires := time.Now().UTC().Add(12 * time.Hour)
	return &Session{
		Id:       id,
		Username: username,
		Expires:  expires.Unix()}
}

func cookieMac(i int64) []byte {
	sessionId := make([]byte, 8)
	binary.BigEndian.PutUint64(sessionId, uint64(i))

	sessionId = append(sessionId,
		hmac.New(sha256.New, hmacKey).Sum(sessionId)...)
	return sessionId
}

func (s *Session) Cookie() *http.Cookie {
	return &http.Cookie{
		Name:    "phosphorus",
		Value:   base64.StdEncoding.EncodeToString(cookieMac(s.Id)),
		Expires: time.Unix(s.Expires, 0)}
}

func (s *Session) Valid() bool {
	return s.Expires > time.Now().UTC().Unix()
}

func ReadCookie(c string) (s *Session, err error) {
	cookie, err := base64.StdEncoding.DecodeString(c)
	if err != nil {
		return
	}

	sessionId := int64(binary.BigEndian.Uint64(cookie[0:8]))
	if !hmac.Equal(cookieMac(sessionId), cookie) {
		err = fmt.Errorf("MAC validation failed")
		return
	}

	s = &Session{Id: sessionId}
	return
}

// resources

var Login = &Resource{
	Path: "/login",
	Get:  LoginForm,
	Post: LoginAttempt}

func LoginForm(r *http.Request, e *Env, m map[string]interface{}) (Response, error) {
	return Static("login"), nil
}

func LoginAttempt(r *http.Request, e *Env, m map[string]interface{}) (Response, error) {
	acct := &Account{Username: r.Form.Get("username")}
	err := db.GetItem(e.Accounts, acct)
	if err != nil {
		return nil, fmt.Errorf("Bad authentication")
		// panic(err)
		// could be an unexpected error or a not found
	}

	err = acct.Authenticate(r.Form.Get("password"))
	if err != nil {
		return nil, fmt.Errorf("Bad authentication")
	}

	m["session"] = NewSession(acct.Username, e.IdGen.SafeId())
	return SeeOther(Dashboard.Path + acct.Username), nil
}

var Enroll = &Resource{
	Path: "/enroll",
	Get:  EnrollForm,
	Post: EnrollAttempt}

func EnrollForm(r *http.Request, e *Env, m map[string]interface{}) (Response, error) {
	return Static("enroll"), nil
}

func EnrollAttempt(r *http.Request, e *Env, m map[string]interface{}) (Response, error) {
	if r.Form.Get("password") != r.Form.Get("confirm") {
		return nil, fmt.Errorf("Password mismatch")
	}

	username := strings.ToLower(r.Form.Get("username"))
	matched, _ := regexp.MatchString("^[a-z][a-z0-9]{2,15}$", username)
	if !matched {
		return nil, fmt.Errorf("Invalid username")
	}

	salt := newSalt()
	password := []byte(r.Form.Get("password"))
	a := &Account{
		Username: r.Form.Get("username"),
		Salt:     b2s(salt),
		Password: b2s(kdf(salt, password))}

	err := db.CreateItem(e.Accounts, a)
	if err != nil {
		return nil, err
	}

	m["session"] = NewSession(username, e.IdGen.SafeId())

	return SeeOther(Dashboard.Path + username), nil
}

// decorator

func Authed(fn Handler) Handler {
	return func(r *http.Request, e *Env, m map[string]interface{}) (resp Response, err error) {
		cookie, err := r.Cookie("phosphorus")
		if err != nil {
			return
		}

		session, err := ReadCookie(cookie.Value)
		if err != nil {
			return
		}

		// session := &Session{Id: sessionId}
		err = db.GetItem(e.Sessions, session)
		if err != nil {
			return
		}

		if !session.Valid() {
			err = fmt.Errorf("old session")
			return
		}

		account := &Account{Username: session.Username}
		err = db.GetItem(e.Accounts, account)
		if err != nil {
			panic(err)
		}
		m["session"] = session
		m["account"] = account
		log.Println("dong")
		return fn(r, e, m)
	}
}

// helper

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
