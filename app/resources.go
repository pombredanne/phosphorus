package app

import (
	"encoding/json"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"html/template"
	"log"
	"net/http"
	"willstclair.com/phosphorus/id"
)

var Templates map[string]*template.Template

var Resources = []*Resource{
	Login,
	Enroll,
	Token,
	Dashboard,
}

type Env struct {
	Auth     *aws.Auth
	Accounts *dynamodb.Table
	Sessions *dynamodb.Table
	IdGen    *id.Generator
}

type Response interface {
	Flush(http.ResponseWriter)
}

type Handler func(*http.Request, *Env, map[string]interface{}) (Response, error)

type Resource struct {
	Path   string
	Env    *Env
	Get    Handler
	Post   Handler
	Put    Handler
	Delete Handler
}

func (c *Resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC: %s\n", r)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal server error")
		}
	}()

	var rep Response
	handler := c.dispatch(r.Method)
	if handler != nil {
		rep, err := handler(r, c.Env, map[string]interface{}{})
		if err != nil {
			rep = &Error{err.Error()}
		}
		rep.Flush(w)
		return
	}

	rep = &Error{"Method not allowed"}
	rep.Flush(w)
}

func (c *Resource) dispatch(method string) (m Handler) {
	switch method {
	case "GET":
		m = c.Get
	case "POST":
		m = c.Post
	case "PUT":
		m = c.Put
	case "DELETE":
		m = c.Delete
	}
	return
}

type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Flush(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func Anonymous(name string, d interface{}) Response {
	return &static{Templates[name+".html"], d}
}

func Static(name string) Response {
	return &static{t: Templates[name+".html"]}
}

type static struct {
	t *template.Template
	d interface{}
}

func (r *static) Flush(w http.ResponseWriter) {
	r.t.ExecuteTemplate(w, "layout", r.d)
}

type redirect struct {
	location string
	code     int
}

func (r *redirect) Flush(w http.ResponseWriter) {
	w.Header().Add("Location", r.location)
	w.WriteHeader(r.code)
}

func SeeOther(loc string) Response {
	return &redirect{loc, 303}
}

func PermanentRedirect(loc string) Response {
	return &redirect{loc, 303}
}

func TemporaryRedirect(loc string) Response {
	return &redirect{loc, 307}
}
