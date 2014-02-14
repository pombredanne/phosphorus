package main

import (
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"willstclair.com/phosphorus/app"
	"willstclair.com/phosphorus/id"
)

var cmdServer = &Command{
	Run:       runServer,
	UsageLine: "server",
	Short:     "server",
}

var (
	serverWebDir string // -web flag
)

func init() {
	cmdServer.Flag.StringVar(&serverWebDir, "web", "", "")
}

func getTemplates() map[string]*template.Template {
	t := make(map[string]*template.Template)
	joined := filepath.Join(serverWebDir, "html/*.html")
	tmplDir, err := filepath.Abs(joined)
	if err != nil {
		panic(err)
	}

	templates, err := filepath.Glob(tmplDir)
	if err != nil {
		panic(err)
	}

	for _, filename := range templates {
		bn := filepath.Base(filename)
		if bn == "layout.html" {
			continue
		}
		t[bn] = template.Must(template.ParseFiles(filename,
			filepath.Join(serverWebDir, "html", "layout.html")))
	}
	return t
}

func runServer(cmd *Command, args []string) {
	app.Templates = getTemplates()

	exp := time.Now().Add(60 * time.Minute)
	auth, err := aws.EnvAuth()
	if err != nil {
		auth, err = aws.GetAuth("", "", "", exp)
		if err != nil {
			panic(err)
		}
	}

	dynamo := &dynamodb.Server{auth, aws.USEast}

	env := &app.Env{
		Sessions: dynTable(dynamo, "session"),
		Accounts: dynTable(dynamo, "account"),
		IdGen:    id.NewGenerator(1),
		Auth:     &auth}

	for _, r := range app.Resources {
		r.Env = env
		http.Handle(r.Path, r)
	}

	// http.Handle("/enroll", &app.Handler{
	// 	Env:      env,
	// 	Methods:  []string{"GET"},
	// 	Template: templates["enroll.html"],
	// 	Run:      app.Enroll})

	// http.HandleFunc("/enroll", app.GetPost(app.Enroll(env)))
	// http.HandleFunc("/login", app.GetPost(app.Login(env)))

	// http.HandleFunc("/u/", app.Get(app.Authed(env, app.Dashboard)))
	// http.HandleFunc("/_form", app.Get(app.Authed(env, app.UploadToken)))

	http.HandleFunc("/_js", jsHandler)
	http.HandleFunc("/_css", cssHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func catGlob(ext string, w http.ResponseWriter) {
	files, err := filepath.Glob(
		filepath.Join(
			serverWebDir,
			fmt.Sprintf("/%s/*.%s", ext, ext)))
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(w, file)
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/javascript")
	catGlob("js", w)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/css")
	catGlob("css", w)
}
